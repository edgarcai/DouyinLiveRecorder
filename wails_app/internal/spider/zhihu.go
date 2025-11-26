package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type ZhihuSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewZhihuSpider() *ZhihuSpider {
	return &ZhihuSpider{}
}

func (s *ZhihuSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *ZhihuSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *ZhihuSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Handle User Profile URL vs Live URL
	var livePageUrl string
	if strings.Contains(targetUrl, "people/") {
		// https://www.zhihu.com/people/username
		parts := strings.Split(targetUrl, "people/")
		if len(parts) > 1 {
			userId := parts[1]
			api := fmt.Sprintf("https://api.zhihu.com/people/%s/profile?profile_new_version=", userId)

			client := &http.Client{}
			req, err := http.NewRequest("GET", api, nil)
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

			if drama, ok := jsonResult["drama"].(map[string]interface{}); ok {
				if livingTheater, ok := drama["living_theater"].(map[string]interface{}); ok {
					if urlStr, ok := livingTheater["theater_url"].(string); ok {
						livePageUrl = urlStr
					}
				}
			}
		}
	} else {
		livePageUrl = targetUrl
	}

	if livePageUrl == "" {
		return nil, fmt.Errorf("failed to find live page url")
	}

	// 2. Extract Web ID
	// https://www.zhihu.com/theater/123456
	parts := strings.Split(strings.Split(livePageUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid live page url")
	}
	webId := parts[len(parts)-1]

	// 3. Fetch Live Page
	client := &http.Client{}
	req, err := http.NewRequest("GET", livePageUrl, nil)
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

	// 4. Extract JSON Data
	re := regexp.MustCompile(`<script id="js-initialData" type="text/json">(.*?)</script>`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find initial data")
	}
	jsonStr := matches[1]

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonResult); err != nil {
		return nil, err
	}

	// 5. Extract Stream URL
	if initialState, ok := jsonResult["initialState"].(map[string]interface{}); ok {
		if theater, ok := initialState["theater"].(map[string]interface{}); ok {
			if theaters, ok := theater["theaters"].(map[string]interface{}); ok {
				if liveData, ok := theaters[webId].(map[string]interface{}); ok {
					if drama, ok := liveData["drama"].(map[string]interface{}); ok {
						if status, ok := drama["status"].(string); ok && status == "END" {
							return nil, fmt.Errorf("live ended")
						}
						// Status seems to be string "END" or maybe int 1? Python code checks `live_status == 1`.
						// Let's check playInfo directly.
						if playInfo, ok := drama["playInfo"].(map[string]interface{}); ok {
							if hlsUrl, ok := playInfo["hlsUrl"].(string); ok && hlsUrl != "" {
								return &StreamInfo{Url: hlsUrl}, nil
							}
							if playUrl, ok := playInfo["playUrl"].(string); ok && playUrl != "" {
								return &StreamInfo{Url: playUrl}, nil
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
