package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MaoerSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewMaoerSpider() *MaoerSpider {
	return &MaoerSpider{}
}

func (s *MaoerSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *MaoerSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *MaoerSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://fm.missevan.com/live/868895007
	parts := strings.Split(targetUrl, "?")[0]
	pathParts := strings.Split(parts, "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := pathParts[len(pathParts)-1]

	// 2. Call API
	apiUrl := fmt.Sprintf("https://fm.missevan.com/api/v2/live/%s", roomId)

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
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
	if info, ok := jsonResult["info"].(map[string]interface{}); ok {
		if room, ok := info["room"].(map[string]interface{}); ok {
			if status, ok := room["status"].(map[string]interface{}); ok {
				if broadcasting, ok := status["broadcasting"].(bool); ok && broadcasting {
					if channel, ok := room["channel"].(map[string]interface{}); ok {
						if flvUrl, ok := channel["flv_pull_url"].(string); ok {
							return &StreamInfo{Url: flvUrl}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
