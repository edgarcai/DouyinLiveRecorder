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

type Room6Spider struct {
	ProxyUrl string
	Cookies  string
}

// NewRoom6Spider 创建 Room6 爬虫实例
func NewRoom6Spider() *Room6Spider {
	return &Room6Spider{}
}

// SetProxy 设置代理地址
func (s *Room6Spider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *Room6Spider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *Room6Spider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://v.6.cn/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	client := &http.Client{}

	// 2. Fetch Page to get real Room ID (ruid)
	// Sometimes the URL ID is not the real room ID
	req, err := http.NewRequest("GET", fmt.Sprintf("https://v.6.cn/%s", roomId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Referer", "https://ios.6.cn/?ver=8.0.3&build=4")
	if s.Cookies != "" {
		req.Header.Set("Cookie", s.Cookies)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	htmlStr := string(bodyBytes)

	// rid: '(\d+)',\n\s+roomid
	re := regexp.MustCompile(`rid: '(.*?)',\n\s+roomid`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) >= 2 {
		roomId = matches[1]
	}

	// 3. Call API
	api := "https://v.6.cn/coop/mobile/index.php?padapi=coop-mobile-inroom.php"
	params := url.Values{}
	params.Set("av", "3.1")
	params.Set("encpass", "")
	params.Set("logiuid", "")
	params.Set("project", "v6iphone")
	params.Set("rate", "1")
	params.Set("rid", "")
	params.Set("ruid", roomId)

	reqApi, err := http.NewRequest("POST", api, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	reqApi.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqApi.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	reqApi.Header.Set("Referer", "https://ios.6.cn/?ver=8.0.3&build=4")
	if s.Cookies != "" {
		reqApi.Header.Set("Cookie", s.Cookies)
	}

	respApi, err := client.Do(reqApi)
	if err != nil {
		return nil, err
	}
	defer respApi.Body.Close()

	bodyBytesApi, err := io.ReadAll(respApi.Body)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(bodyBytesApi, &jsonResult); err != nil {
		return nil, err
	}

	// 4. Extract Stream URL
	if content, ok := jsonResult["content"].(map[string]interface{}); ok {
		if liveInfo, ok := content["liveinfo"].(map[string]interface{}); ok {
			if flvTitle, ok := liveInfo["flvtitle"].(string); ok && flvTitle != "" {
				return &StreamInfo{Url: fmt.Sprintf("https://wlive.6rooms.com/httpflv/%s.flv", flvTitle)}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
