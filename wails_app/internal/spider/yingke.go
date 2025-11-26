package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type YingkeSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewYingkeSpider 创建 Yingke 爬虫实例
func NewYingkeSpider() *YingkeSpider {
	return &YingkeSpider{}
}

// SetProxy 设置代理地址
func (s *YingkeSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *YingkeSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *YingkeSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Params
	// https://www.inke.cn/live_share.html?uid=123&id=456
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}
	uid := u.Query().Get("uid")
	liveId := u.Query().Get("id")

	if uid == "" || liveId == "" {
		return nil, fmt.Errorf("invalid url: missing uid or id")
	}

	// 2. Call API
	params := url.Values{}
	params.Set("uid", uid)
	params.Set("id", liveId)
	params.Set("_t", fmt.Sprintf("%d", time.Now().Unix()))

	api := fmt.Sprintf("https://webapi.busi.inke.cn/web/live_share_pc?%s", params.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Referer", "https://www.inke.cn/")
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
		if status, ok := data["status"].(float64); ok && status == 1 {
			if liveInfo, ok := data["live_info"].([]interface{}); ok && len(liveInfo) > 0 {
				if info, ok := liveInfo[0].(map[string]interface{}); ok {
					if streamAddr, ok := info["stream_addr"].(string); ok && streamAddr != "" {
						return &StreamInfo{Url: streamAddr}, nil
					}
				}
			}
			// Fallback to file_info -> media_url if stream_addr not found (not shown in Python code but common)
			// Python code uses data['data']['file_info']['media_url'] as fallback?
			// Wait, Python code:
			/*
			   if live_status == 1:
			       media_url = json_data['data']['file_info']['media_url']
			       result |= {
			           'is_live': True,
			           'm3u8_url': media_url,
			           'record_url': media_url
			       }
			*/
			// Ah, Python code uses `file_info` -> `media_url`.
			if fileInfo, ok := data["file_info"].(map[string]interface{}); ok {
				if mediaUrl, ok := fileInfo["media_url"].(string); ok && mediaUrl != "" {
					return &StreamInfo{Url: mediaUrl}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
