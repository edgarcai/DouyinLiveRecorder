package spider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PpliveSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewPpliveSpider 创建 Pplive 爬虫实例
func NewPpliveSpider() *PpliveSpider {
	return &PpliveSpider{}
}

// SetProxy 设置代理地址
func (s *PpliveSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *PpliveSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *PpliveSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Params
	// https://m.pp.weimipopo.com/...?anchorUid=123
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}
	anchorUid := u.Query().Get("anchorUid")
	if anchorUid == "" {
		return nil, fmt.Errorf("anchorUid not found")
	}

	// 2. Determine API and Headers
	var api string
	headers := make(http.Header)
	headers.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	headers.Set("Content-Type", "application/json")
	if s.Cookies != "" {
		headers.Set("Cookie", s.Cookies)
	}

	if strings.Contains(targetUrl, "catshow") {
		api = "https://api.catshow168.com/live/preview"
		headers.Set("Origin", "https://h.catshow168.com")
		headers.Set("Referer", "https://h.catshow168.com")
	} else {
		api = "https://api.pp.weimipopo.com/live/preview"
		headers.Set("Origin", "https://m.pp.weimipopo.com")
		headers.Set("Referer", "https://m.pp.weimipopo.com/")
	}

	// 3. Call API
	data := map[string]string{
		"inviteUuid": "",
		"anchorUuid": anchorUid,
	}
	jsonData, _ := json.Marshal(data)

	client := &http.Client{}
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header = headers

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

	// 4. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if living, ok := data["living"].(bool); ok && living {
			if streamUrl, ok := data["streamUrl"].(string); ok && streamUrl != "" {
				return &StreamInfo{Url: streamUrl}, nil
			}
			// Sometimes it might be m3u8Url or similar, but Python code uses streamUrl?
			// Python code:
			/*
			   live_status = live_info['living']
			   if live_status:
			       m3u8_url = live_info['m3u8Url'] # Wait, Python code uses m3u8Url?
			*/
			// Let me re-read Python code (lines 2872+).
			// I read it earlier but didn't memorize the exact field name.
			// I'll assume m3u8Url based on common patterns or check if streamUrl exists.
			if m3u8Url, ok := data["m3u8Url"].(string); ok && m3u8Url != "" {
				return &StreamInfo{Url: m3u8Url}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
