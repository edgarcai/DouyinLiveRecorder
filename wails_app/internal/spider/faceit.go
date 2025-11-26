package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type FaceitSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewFaceitSpider 创建 Faceit 爬虫实例
func NewFaceitSpider() *FaceitSpider {
	return &FaceitSpider{}
}

// SetProxy 设置代理地址
func (s *FaceitSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *FaceitSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *FaceitSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Nickname
	// https://www.faceit.com/zh/players/qpjzz/stream
	re := regexp.MustCompile(`/players/(.*?)/stream`)
	matches := re.FindStringSubmatch(targetUrl)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to extract nickname")
	}
	nickname := matches[1]

	// 2. Get User ID
	api := fmt.Sprintf("https://www.faceit.com/api/users/v1/nicknames/%s", nickname)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Referer", targetUrl)
	req.Header.Set("faceit-referer", "web-next")
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

	var userId string
	if payload, ok := jsonResult["payload"].(map[string]interface{}); ok {
		if id, ok := payload["id"].(string); ok {
			userId = id
		}
	}

	if userId == "" {
		return nil, fmt.Errorf("user id not found")
	}

	// 3. Get Stream Info
	api2 := fmt.Sprintf("https://www.faceit.com/api/stream/v1/streamings?userId=%s", userId)
	req2, err := http.NewRequest("GET", api2, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req2.Header.Set("Referer", targetUrl)
	req2.Header.Set("faceit-referer", "web-next")
	if s.Cookies != "" {
		req2.Header.Set("Cookie", s.Cookies)
	}

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	bodyBytes2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}

	var jsonResult2 map[string]interface{}
	if err := json.Unmarshal(bodyBytes2, &jsonResult2); err != nil {
		return nil, err
	}

	if payload, ok := jsonResult2["payload"].([]interface{}); ok && len(payload) > 0 {
		if platformInfo, ok := payload[0].(map[string]interface{}); ok {
			platform, _ := platformInfo["platform"].(string)
			anchorId, _ := platformInfo["platformId"].(string)

			if platform == "twitch" {
				// Delegate to TwitchSpider
				twitchUrl := fmt.Sprintf("https://www.twitch.tv/%s", anchorId)
				twitchSpider := NewTwitchSpider()
				twitchSpider.SetProxy(s.ProxyUrl)
				twitchSpider.SetCookies(s.Cookies)
				return twitchSpider.GetStreamUrl(twitchUrl)
			}
		}
	}

	return nil, fmt.Errorf("stream not found or platform not supported")
}
