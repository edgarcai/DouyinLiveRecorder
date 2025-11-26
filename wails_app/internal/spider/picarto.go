package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type PicartoSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewPicartoSpider 创建 Picarto 爬虫实例
func NewPicartoSpider() *PicartoSpider {
	return &PicartoSpider{}
}

// SetProxy 设置代理地址
func (s *PicartoSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *PicartoSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *PicartoSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Anchor ID
	// https://picarto.tv/username
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	anchorId := parts[len(parts)-1]

	// 2. Call API
	api := fmt.Sprintf("https://ptvintern.picarto.tv/api/channel/detail/%s", anchorId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
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
	if channel, ok := jsonResult["channel"].(map[string]interface{}); ok {
		if online, ok := channel["online"].(bool); ok && online {
			if name, ok := channel["name"].(string); ok {
				m3u8Url := fmt.Sprintf("https://1-edge1-us-newyork.picarto.tv/stream/hls/golive+%s/index.m3u8", name)
				return &StreamInfo{Url: m3u8Url}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
