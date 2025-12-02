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
	// 1. 处理 xhslink.com 重定向
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
		// 模仿 Python 中的 Headers
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

	// 2. 获取 HTML
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

	// 3. 从 window.__INITIAL_STATE__ 解析 JSON
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

	// 4. 提取流 URL
	if liveStream, ok := jsonData["liveStream"].(map[string]interface{}); ok {
		if liveStatus, ok := liveStream["liveStatus"].(string); ok && liveStatus == "success" {
			if roomData, ok := liveStream["roomData"].(map[string]interface{}); ok {
				if roomInfo, ok := roomData["roomInfo"].(map[string]interface{}); ok {
					if liveLink, ok := roomInfo["deeplink"].(string); ok {
						// 提取 flvUrl 参数
						// xhslink://live/room?flvUrl=...
						reFlv := regexp.MustCompile(`flvUrl=(.*?)&`)
						flvMatches := reFlv.FindStringSubmatch(liveLink)
						if len(flvMatches) < 2 {
							// 尝试字符串结尾
							reFlvEnd := regexp.MustCompile(`flvUrl=(.*?)$`)
							flvMatches = reFlvEnd.FindStringSubmatch(liveLink)
						}

						if len(flvMatches) >= 2 {
							flvUrl := flvMatches[1]
							// 如果需要，解码 URL，但通常它是原始的
							// 逻辑来自 Python:
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
