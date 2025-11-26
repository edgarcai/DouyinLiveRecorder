package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ChangliaoSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewChangliaoSpider 创建 Changliao 爬虫实例
func NewChangliaoSpider() *ChangliaoSpider {
	return &ChangliaoSpider{}
}

// SetProxy 设置代理地址
func (s *ChangliaoSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *ChangliaoSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *ChangliaoSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://wap.tlclw.com/phone/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Call API
	params := url.Values{}
	params.Set("roomidx", roomId)
	params.Set("currentUrl", fmt.Sprintf("https://wap.tlclw.com/%s", roomId))

	api := fmt.Sprintf("https://wap.tlclw.com/api/ui/room/v1.0.0/live.ashx?%s", params.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Set("Referer", fmt.Sprintf("https://wap.tlclw.com/phone/%s?promoters=0", roomId))
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

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
		return nil, err
	}

	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if roomInfo, ok := data["roomInfo"].(map[string]interface{}); ok {
			if liveStat, ok := roomInfo["live_stat"].(float64); ok && liveStat == 1 {
				if videoUrl, ok := roomInfo["videoUrl"].(string); ok && videoUrl != "" {
					return &StreamInfo{Url: videoUrl}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
