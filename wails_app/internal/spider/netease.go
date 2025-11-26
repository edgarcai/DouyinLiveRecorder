package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type NeteaseSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewNeteaseSpider() *NeteaseSpider {
	return &NeteaseSpider{}
}

func (s *NeteaseSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *NeteaseSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *NeteaseSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Fetch HTML
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Referer", "https://cc.163.com/")
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

	// 2. Extract JSON from __NEXT_DATA__
	re := regexp.MustCompile(`<script id="__NEXT_DATA__" .* crossorigin="anonymous">(.*?)</script></body>`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find __NEXT_DATA__")
	}
	jsonStr := matches[1]

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return nil, err
	}

	// 3. Extract Stream URL
	// props -> pageProps -> roomInfoInitData -> live -> quickplay -> resolution
	if props, ok := jsonData["props"].(map[string]interface{}); ok {
		if pageProps, ok := props["pageProps"].(map[string]interface{}); ok {
			if roomInfo, ok := pageProps["roomInfoInitData"].(map[string]interface{}); ok {
				if live, ok := roomInfo["live"].(map[string]interface{}); ok {
					if quickplay, ok := live["quickplay"].(map[string]interface{}); ok {
						if resolution, ok := quickplay["resolution"].(map[string]interface{}); ok {
							// Prefer blueray > ultra > high > standard
							priorities := []string{"blueray", "ultra", "high", "standard"}
							for _, quality := range priorities {
								if streamData, ok := resolution[quality].(map[string]interface{}); ok {
									if cdn, ok := streamData["cdn"].(map[string]interface{}); ok {
										// Iterate CDNs and pick first valid
										for _, urlVal := range cdn {
											if urlStr, ok := urlVal.(string); ok {
												return &StreamInfo{Url: urlStr}, nil
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
