package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LangliveSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewLangliveSpider 创建 Langlive 爬虫实例
func NewLangliveSpider() *LangliveSpider {
	return &LangliveSpider{}
}

// SetProxy 设置代理地址
func (s *LangliveSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *LangliveSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *LangliveSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://www.lang.live/room/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Call API
	api := fmt.Sprintf("https://api.lang.live/langweb/v1/room/liveinfo?room_id=%s", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Origin", "https://www.lang.live")
	req.Header.Set("Referer", "https://www.lang.live/")
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

	// 3. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if liveInfo, ok := data["live_info"].(map[string]interface{}); ok {
			if liveStatus, ok := liveInfo["live_status"].(float64); ok && liveStatus == 1 {
				if liveUrl, ok := liveInfo["liveurl"].(string); ok && liveUrl != "" {
					return &StreamInfo{Url: liveUrl}, nil
				}
				if liveUrlHls, ok := liveInfo["liveurl_hls"].(string); ok && liveUrlHls != "" {
					return &StreamInfo{Url: liveUrlHls}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
