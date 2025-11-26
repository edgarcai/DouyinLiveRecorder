package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Live17Spider struct {
	ProxyUrl string
	Cookies  string
}

// NewLive17Spider 创建 17Live 爬虫实例
func NewLive17Spider() *Live17Spider {
	return &Live17Spider{}
}

// SetProxy 设置代理地址
func (s *Live17Spider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *Live17Spider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *Live17Spider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://17.live/en/live/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Check Live Status
	client := &http.Client{}
	api := fmt.Sprintf("https://wap-api.17app.co/api/v1/lives/%s/viewers/alive", roomId)

	dataMap := map[string]string{
		"liveStreamID": roomId,
	}
	dataJson, _ := json.Marshal(dataMap)

	req, err := http.NewRequest("POST", api, strings.NewReader(string(dataJson)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Origin", "https://17.live")
	req.Header.Set("Referer", "https://17.live/")
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
	if status, ok := jsonResult["status"].(float64); ok && status == 2 {
		if pullURLsInfo, ok := jsonResult["pullURLsInfo"].(map[string]interface{}); ok {
			if rtmpURLs, ok := pullURLsInfo["rtmpURLs"].([]interface{}); ok && len(rtmpURLs) > 0 {
				if rtmp0, ok := rtmpURLs[0].(map[string]interface{}); ok {
					if urlHighQuality, ok := rtmp0["urlHighQuality"].(string); ok {
						return &StreamInfo{Url: urlHighQuality}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
