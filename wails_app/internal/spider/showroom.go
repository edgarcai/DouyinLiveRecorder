package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type ShowRoomSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewShowRoomSpider 创建 ShowRoom 爬虫实例
func NewShowRoomSpider() *ShowRoomSpider {
	return &ShowRoomSpider{}
}

// SetProxy 设置代理地址
func (s *ShowRoomSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *ShowRoomSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *ShowRoomSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	var roomId string
	if strings.Contains(targetUrl, "/room/profile") {
		// https://www.showroom-live.com/room/profile?room_id=123456
		parts := strings.Split(targetUrl, "room_id=")
		if len(parts) > 1 {
			roomId = parts[1]
		}
	} else {
		// https://www.showroom-live.com/roomname
		client := &http.Client{}
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
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

		re := regexp.MustCompile(`href="/room/profile\?room_id=(.*?)"`)
		matches := re.FindStringSubmatch(htmlStr)
		if len(matches) >= 2 {
			roomId = matches[1]
		}
	}

	if roomId == "" {
		return nil, fmt.Errorf("failed to extract room id")
	}

	// 2. Get Live Info
	infoApi := fmt.Sprintf("https://www.showroom-live.com/api/live/live_info?room_id=%s", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", infoApi, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
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

	liveStatus, ok := jsonResult["live_status"].(float64)
	if !ok || liveStatus != 2 {
		return nil, fmt.Errorf("stream not found or live ended")
	}

	// 3. Get Streaming URL
	webApi := fmt.Sprintf("https://www.showroom-live.com/api/live/streaming_url?room_id=%s&abr_available=1", roomId)
	reqWeb, err := http.NewRequest("GET", webApi, nil)
	if err != nil {
		return nil, err
	}
	reqWeb.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	reqWeb.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	if s.Cookies != "" {
		reqWeb.Header.Set("Cookie", s.Cookies)
	}

	respWeb, err := client.Do(reqWeb)
	if err != nil {
		return nil, err
	}
	defer respWeb.Body.Close()

	webBody, err := io.ReadAll(respWeb.Body)
	if err != nil {
		return nil, err
	}

	var jsonWeb map[string]interface{}
	if err := json.Unmarshal(webBody, &jsonWeb); err != nil {
		return nil, err
	}

	if streamingUrlList, ok := jsonWeb["streaming_url_list"].([]interface{}); ok {
		// Iterate and find hls
		for _, item := range streamingUrlList {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, ok := itemMap["type"].(string); ok && itemType == "hls" {
					if urlStr, ok := itemMap["url"].(string); ok {
						return &StreamInfo{Url: urlStr}, nil
					}
				}
			}
		}
		// Fallback to any url if hls not found
		if len(streamingUrlList) > 0 {
			if itemMap, ok := streamingUrlList[0].(map[string]interface{}); ok {
				if urlStr, ok := itemMap["url"].(string); ok {
					return &StreamInfo{Url: urlStr}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
