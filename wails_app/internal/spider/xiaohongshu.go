package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type XiaohongshuSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewXiaohongshuSpider() *XiaohongshuSpider {
	return &XiaohongshuSpider{}
}

func (s *XiaohongshuSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *XiaohongshuSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *XiaohongshuSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Handle xhslink.com redirection
	if strings.Contains(targetUrl, "xhslink.com") {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		// Mimic headers from Python
		req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
		req.Header.Set("xy-common-params", "platform=iOS&sid=session.1722166379345546829388")
		req.Header.Set("referer", "https://app.xhs.cn/")
		if s.Cookies != "" {
			req.Header.Set("Cookie", s.Cookies)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 301 || resp.StatusCode == 302 {
			targetUrl = resp.Header.Get("Location")
		}
	}

	// 2. Fetch HTML
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("xy-common-params", "platform=iOS&sid=session.1722166379345546829388")
	req.Header.Set("referer", "https://app.xhs.cn/")
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

	// 3. Parse JSON from window.__INITIAL_STATE__
	re := regexp.MustCompile(`<script>window.__INITIAL_STATE__=(.*?)</script>`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find __INITIAL_STATE__")
	}

	jsonStr := matches[1]
	jsonStr = strings.ReplaceAll(jsonStr, "undefined", "null")

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 4. Extract Stream URL
	if liveStream, ok := jsonData["liveStream"].(map[string]interface{}); ok {
		if liveStatus, ok := liveStream["liveStatus"].(string); ok && liveStatus == "success" {
			if roomData, ok := liveStream["roomData"].(map[string]interface{}); ok {
				if roomInfo, ok := roomData["roomInfo"].(map[string]interface{}); ok {
					if liveLink, ok := roomInfo["deeplink"].(string); ok {
						// Extract flvUrl param
						// xhslink://live/room?flvUrl=...
						reFlv := regexp.MustCompile(`flvUrl=(.*?)&`)
						flvMatches := reFlv.FindStringSubmatch(liveLink)
						if len(flvMatches) < 2 {
							// Try end of string
							reFlvEnd := regexp.MustCompile(`flvUrl=(.*?)$`)
							flvMatches = reFlvEnd.FindStringSubmatch(liveLink)
						}

						if len(flvMatches) >= 2 {
							flvUrl := flvMatches[1]
							// Decode URL if needed, but usually it's raw
							// Logic from Python:
							// room_id = flv_url.split('live/')[1].split('.')[0]
							// flv_url = f"http://live-source-play.xhscdn.com/live/{room_id}.flv"
							if strings.Contains(flvUrl, "live/") {
								parts := strings.Split(flvUrl, "live/")
								if len(parts) > 1 {
									roomId := strings.Split(parts[1], ".")[0]
									finalUrl := fmt.Sprintf("http://live-source-play.xhscdn.com/live/%s.flv", roomId)
									return &StreamInfo{Url: finalUrl}, nil
								}
							}
							return &StreamInfo{Url: flvUrl}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
