package spider

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type DouyuSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewDouyuSpider() *DouyuSpider {
	return &DouyuSpider{
		Client: &http.Client{},
	}
}

func (d *DouyuSpider) SetProxy(proxyUrl string) {
	d.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			d.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (d *DouyuSpider) SetCookies(cookies string) {
	d.Cookies = cookies
}

func (d *DouyuSpider) GetStreamUrl(url string) (*StreamInfo, error) {
	// 逻辑移植自 get_douyu_stream_data
	// 1. 获取房间 ID
	rid := ""
	reRid := regexp.MustCompile(`rid=(\d+)`)
	matches := reRid.FindStringSubmatch(url)
	if len(matches) > 1 {
		rid = matches[1]
	}

	if rid == "" {
		parts := strings.Split(url, "douyu.com/")
		if len(parts) > 1 {
			rid = strings.Split(parts[1], "?")[0]
		}
	}

	if rid == "" {
		return nil, fmt.Errorf("cannot find room id")
	}

	// 2. 获取信息（检查是否直播）
	// https://www.douyu.com/betard/{rid}
	infoUrl := fmt.Sprintf("https://www.douyu.com/betard/%s", rid)
	req, _ := http.NewRequest("GET", infoUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var infoData map[string]interface{}
	if err := json.Unmarshal(body, &infoData); err != nil {
		return nil, err
	}

	room, ok := infoData["room"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("room info not found")
	}

	showStatus, _ := room["show_status"].(float64)
	videoLoop, _ := room["videoLoop"].(float64)
	if showStatus != 1 || videoLoop != 0 {
		return nil, fmt.Errorf("room is not live")
	}

	// 3. 获取流数据（签名逻辑）
	// python 代码使用 execjs 运行一些 JS 逻辑进行签名。
	// 如果没有 JS 引擎，这很难 1:1 移植。
	// 但是，对于斗鱼，通常移动 API 或更简单的 API 在没有复杂 JS 的情况下也可以工作，只要我们有正确的 headers/did。
	// Python 代码使用: https://m.douyu.com/3125893?rid=...
	// 并提取 JS 函数 `ub98484234`。

	// 简化方法：尝试使用可能更宽松或使用固定 DID 的移动 Web API。
	// 实际上，Python 代码在 `get_token_js` 中实现了完整的 JS 逆向。
	// 它提取一个函数，修改它，并运行它。
	// 由于我们无法运行 JS，除非我们找到纯 Go 实现或不同的 API，否则我们可能会陷入困境。

	// 替代方案：直接使用 H5 移动 API，有时会给出没有复杂签名的链接？
	// 或者尝试直接从页面源代码中获取 HLS 链接（如果可用）。

	// 让我们尝试通常更容易的 `h5_m_station` API。
	// POST https://playweb.douyucdn.cn/lapi/live/hlsH5Preview/{rid}

	// timestamp := time.Now().UnixNano() / 1e6
	// did := "10000000000000000000000000001501"
	// auth	sign := douyuMd5Sum(fmt.Sprintf("%s%s%s%s", roomId, did, timestamp, "2501"))

	// 注意：如果没有 JS，斗鱼签名众所周知很难维护。
	// 对于此任务，我将实现一个占位符，警告有关 JS 要求，
	// 或者尝试一个已知的更简单的端点。

	// 让我们尝试从 m.douyu.com 页面源代码中获取，有时 JSON 中包含直播 url。
	mobileUrl := fmt.Sprintf("https://m.douyu.com/%s", rid)
	reqM, _ := http.NewRequest("GET", mobileUrl, nil)
	reqM.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1")
	respM, err := d.Client.Do(reqM)
	if err != nil {
		return nil, err
	}
	defer respM.Body.Close()
	bodyM, _ := io.ReadAll(respM.Body)
	htmlM := string(bodyM)

	// 在 JSON 中查找 "room_url" 或类似内容
	reRoomUrl := regexp.MustCompile(`"url":"(http.*?\.m3u8.*?)"`)
	matchesM := reRoomUrl.FindStringSubmatch(htmlM)
	if len(matchesM) > 1 {
		return &StreamInfo{Url: strings.Replace(matchesM[1], "\\", "", -1)}, nil
	}

	return nil, fmt.Errorf("douyu signature logic requires JS engine, fallback failed")
}

func douyuMd5Sum(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
