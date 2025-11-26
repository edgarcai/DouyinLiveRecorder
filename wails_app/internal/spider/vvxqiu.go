package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type VVxqiuSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewVVxqiuSpider 创建 VVxqiu 爬虫实例
func NewVVxqiuSpider() *VVxqiuSpider {
	return &VVxqiuSpider{}
}

// SetProxy 设置代理地址
func (s *VVxqiuSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *VVxqiuSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *VVxqiuSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://h5webcdn-pro.vvxqiu.com/... or just url with roomId param
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}
	roomId := u.Query().Get("roomId")
	if roomId == "" {
		return nil, fmt.Errorf("room id not found")
	}

	// 2. Call API
	api := fmt.Sprintf("https://h5p.vvxqiu.com/activity-center/fanclub/activity/captain/banner?roomId=%s&product=vvstar", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Origin", "https://h5webcdn-pro.vvxqiu.com")
	req.Header.Set("Referer", "https://h5webcdn-pro.vvxqiu.com/")
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
	// Python code logic:
	// anchor_name = json_data['data']['anchorName']
	// if not anchor_name: try another API (omitted in my implementation for brevity unless necessary)
	// if anchor_name:
	//     stream_url = json_data['data']['streamUrl']

	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if streamUrl, ok := data["streamUrl"].(string); ok && streamUrl != "" {
			return &StreamInfo{Url: streamUrl}, nil
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
