package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type BigoSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewBigoSpider() *BigoSpider {
	return &BigoSpider{}
}

func (s *BigoSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *BigoSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *BigoSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	var roomId string
	if !strings.Contains(targetUrl, "bigo.tv") {
		// Handle short links or other domains if any, fetch to get real URL
		// Simplified: assume bigo.tv or fetch first
		client := &http.Client{}
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
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

		reUrl := regexp.MustCompile(`<meta data-n-head="ssr" data-hid="al:web:url" property="al:web:url" content="(.*?)">`)
		matches := reUrl.FindStringSubmatch(htmlStr)
		if len(matches) >= 2 {
			// content="http://www.bigo.tv/cn/123456?..."
			// extract 123456
			parts := strings.Split(matches[1], "&h=")
			if len(parts) > 1 {
				roomId = parts[len(parts)-1]
			}
		}
	} else {
		if strings.Contains(targetUrl, "&h=") {
			parts := strings.Split(targetUrl, "&h=")
			roomId = parts[len(parts)-1]
		} else {
			// https://www.bigo.tv/123456
			u, err := url.Parse(targetUrl)
			if err == nil {
				pathParts := strings.Split(u.Path, "/")
				if len(pathParts) > 0 {
					roomId = pathParts[len(pathParts)-1]
				}
			}
		}
	}

	if roomId == "" {
		return nil, fmt.Errorf("failed to extract room id")
	}

	// 2. Call API
	apiUrl := "https://ta.bigo.tv/official_website/studio/getInternalStudioInfo"
	data := url.Values{}
	data.Set("siteId", roomId)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", "https://www.bigo.tv/")
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

	// 3. Extract HLS URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if hlsSrc, ok := data["hls_src"].(string); ok && hlsSrc != "" {
			return &StreamInfo{Url: hlsSrc}, nil
		}
	}

	return nil, fmt.Errorf("stream not found")
}
