package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type BluedSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewBluedSpider() *BluedSpider {
	return &BluedSpider{}
}

func (s *BluedSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *BluedSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *BluedSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Fetch HTML
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
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
	htmlStr := string(bodyBytes)

	// 2. Extract JSON
	re := regexp.MustCompile(`decodeURIComponent\("(.*?)"\)\),window\.Promise`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find JSON data")
	}

	jsonStr, err := url.QueryUnescape(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to unescape JSON: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 3. Extract Stream URL
	if userInfo, ok := jsonData["userInfo"].(map[string]interface{}); ok {
		if onLive, ok := userInfo["onLive"].(bool); ok && onLive {
			if liveInfo, ok := jsonData["liveInfo"].(map[string]interface{}); ok {
				if liveUrl, ok := liveInfo["liveUrl"].(string); ok {
					return &StreamInfo{Url: liveUrl}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
