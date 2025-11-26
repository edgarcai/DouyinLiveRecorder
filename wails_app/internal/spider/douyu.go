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
	// Logic ported from get_douyu_stream_data
	// 1. Get Room ID
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

	// 2. Get Info (to check if live)
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

	// 3. Get Stream Data (Sign logic)
	// The python code uses execjs to run some JS logic for signature.
	// This is complex to port 1:1 without a JS engine.
	// However, for Douyu, often the mobile API or a simpler API works without complex JS if we have the right headers/did.
	// Python code uses: https://m.douyu.com/3125893?rid=...
	// And extracts a JS function `ub98484234`.

	// Simplified approach: Try to use the mobile web API which might be more lenient or use a fixed DID.
	// Actually, the Python code implements a full JS reversal in `get_token_js`.
	// It extracts a function, modifies it, and runs it.
	// Since we can't run JS, we might be stuck unless we find a pure Go implementation or a different API.

	// ALTERNATIVE: Use the H5 mobile API directly which sometimes gives a link without complex sign?
	// Or try to fetch the HLS link directly if available in the page source.

	// Let's try the `h5_m_station` API which is often easier.
	// POST https://playweb.douyucdn.cn/lapi/live/hlsH5Preview/{rid}

	// timestamp := time.Now().UnixNano() / 1e6
	// did := "10000000000000000000000000001501"
	// auth	sign := douyuMd5Sum(fmt.Sprintf("%s%s%s%s", roomId, did, timestamp, "2501"))

	// Note: Douyu signature is notoriously hard to maintain without JS.
	// For this task, I will implement a placeholder that warns about JS requirement,
	// OR try a known simpler endpoint.

	// Let's try to fetch from the m.douyu.com page source which sometimes has the live url in JSON.
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

	// Look for "room_url" or similar in JSON
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
