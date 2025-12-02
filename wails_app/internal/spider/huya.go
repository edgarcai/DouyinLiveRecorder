package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type HuyaSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewHuyaSpider() *HuyaSpider {
	return &HuyaSpider{
		Client: &http.Client{},
	}
}

func (h *HuyaSpider) SetProxy(proxyUrl string) {
	h.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			h.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (h *HuyaSpider) SetCookies(cookies string) {
	h.Cookies = cookies
}

func (h *HuyaSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 逻辑移植自 get_huya_app_stream_url（比 web 更简单且更健壮）
	// 提取 room_id
	// url 格式: https://www.huya.com/123456
	parts := strings.Split(targetUrl, "/")
	if len(parts) < 2 { // Changed from < 1 to < 2
		return nil, fmt.Errorf("invalid huya url")
	}
	roomId := parts[len(parts)-1]
	roomId = strings.Split(roomId, "?")[0]

	// Headers
	headers := map[string]string{
		"User-Agent":      "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))",
		"xweb_xhr":        "1",
		"referer":         "https://servicewechat.com/wx74767bf0b684f7d3/301/page-frame.html",
		"accept-language": "zh-CN,zh;q=0.9",
	}

	// 如果 room_id 包含字母，我们可能需要解析它，但让我们先尝试直接 API 或暂时假设为数字。
	// python 代码通过先获取页面来处理字母数字 room_id。
	// 如果不是数字，让我们实现解析。
	isNumeric := regexp.MustCompile(`^\d+$`).MatchString(roomId)
	if !isNumeric {
		// 解析房间 ID
		req, _ := http.NewRequest("GET", targetUrl, nil)
		req.Header.Set("User-Agent", headers["User-Agent"])
		resp, err := h.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		htmlStr := string(body)

		re := regexp.MustCompile(`ProfileRoom":(.*?),"sPrivateHost`)
		matches := re.FindStringSubmatch(htmlStr)
		if len(matches) > 1 {
			roomId = matches[1]
		} else {
			return nil, fmt.Errorf("failed to resolve alphanumeric room id")
		}
	}

	// API 请求
	params := url.Values{}
	params.Set("m", "Live")
	params.Set("do", "profileRoom")
	params.Set("roomid", roomId)
	params.Set("showSecret", "1")

	apiUrl := fmt.Sprintf("https://mp.huya.com/cache.php?%s", params.Encode())
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse stream info: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data not found")
	}

	realLiveStatus, ok := data["realLiveStatus"].(string)
	if !ok || realLiveStatus != "ON" {
		return nil, fmt.Errorf("stream is not live")
	}

	streamObj, ok := data["stream"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stream info not found")
	}

	baseSteamInfoList, ok := streamObj["baseSteamInfoList"].([]interface{})
	if !ok || len(baseSteamInfoList) == 0 {
		return nil, fmt.Errorf("stream list empty")
	}

	// 优先级: TX > HW > HS > AL
	priorityOrder := []string{"TX", "HW", "HS", "AL"}
	var selectedFlvUrl string
	var selectedCdnType string

	// 解析所有流
	type StreamData struct { // 重命名以避免与 StreamInfo 结构体冲突
		CdnType string
		FlvUrl  string
	}
	var streams []StreamData

	for _, item := range baseSteamInfoList {
		info := item.(map[string]interface{})
		cdnType := info["sCdnType"].(string)
		streamName := info["sStreamName"].(string)
		sFlvUrl := info["sFlvUrl"].(string)
		flvAntiCode := info["sFlvAntiCode"].(string)

		flvUrl := fmt.Sprintf("%s/%s.flv?%s", sFlvUrl, streamName, flvAntiCode)
		streams = append(streams, StreamData{CdnType: cdnType, FlvUrl: flvUrl})
	}

	// 选择最佳
	for _, cdn := range priorityOrder {
		for _, s := range streams {
			if s.CdnType == cdn {
				selectedFlvUrl = s.FlvUrl
				selectedCdnType = cdn
				break
			}
		}
		if selectedFlvUrl != "" {
			break
		}
	}

	if selectedFlvUrl == "" && len(streams) > 0 {
		selectedFlvUrl = streams[0].FlvUrl
		selectedCdnType = streams[0].CdnType
	}

	if selectedFlvUrl == "" {
		return nil, fmt.Errorf("no valid stream found")
	}

	// 修复 URL 协议
	if !strings.HasPrefix(selectedFlvUrl, "http") {
		// python 代码通过 :// 分割并添加 https://，假设它可能缺少协议或 http
		// 但 sFlvUrl 通常带有 http。让我们确保 https。
		// Python: flv_url = 'https://' + selected_flv_url.split('://')[1]
		parts := strings.Split(selectedFlvUrl, "://")
		if len(parts) > 1 {
			selectedFlvUrl = "https://" + parts[1]
		}
	}

	// TX 特定修复
	if selectedCdnType == "TX" {
		selectedFlvUrl = strings.Replace(selectedFlvUrl, "&ctype=tars_mp", "&ctype=huya_webh5", -1)
		selectedFlvUrl = strings.Replace(selectedFlvUrl, "&fs=bhct", "&fs=bgct", -1)
	}

	return &StreamInfo{Url: selectedFlvUrl}, nil
}
