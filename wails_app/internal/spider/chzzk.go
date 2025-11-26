package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CHZZKSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewCHZZKSpider 创建 CHZZK 爬虫实例
func NewCHZZKSpider() *CHZZKSpider {
	return &CHZZKSpider{}
}

// SetProxy 设置代理地址
func (s *CHZZKSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *CHZZKSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *CHZZKSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://chzzk.naver.com/live/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Call API
	playApi := fmt.Sprintf("https://api.chzzk.naver.com/service/v3/channels/%s/live-detail", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", playApi, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Origin", "https://chzzk.naver.com")
	req.Header.Set("Referer", targetUrl)
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
	if content, ok := jsonResult["content"].(map[string]interface{}); ok {
		if status, ok := content["status"].(string); ok && status == "OPEN" {
			if livePlaybackJsonStr, ok := content["livePlaybackJson"].(string); ok {
				var livePlaybackJson map[string]interface{}
				if err := json.Unmarshal([]byte(livePlaybackJsonStr), &livePlaybackJson); err != nil {
					return nil, err
				}

				if media, ok := livePlaybackJson["media"].([]interface{}); ok && len(media) > 0 {
					if media0, ok := media[0].(map[string]interface{}); ok {
						if path, ok := media0["path"].(string); ok {
							return &StreamInfo{Url: path}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
