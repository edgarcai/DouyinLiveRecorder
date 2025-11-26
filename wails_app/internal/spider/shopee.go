package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ShopeeSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewShopeeSpider() *ShopeeSpider {
	return &ShopeeSpider{}
}

func (s *ShopeeSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *ShopeeSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *ShopeeSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Handle Redirects and Extract Info
	// https://live.shopee.sg/share?from=live&session=802458&share_user_id=

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}

	// If not already a direct live url, fetch to get redirect
	if !strings.Contains(targetUrl, "live.shopee") && !strings.Contains(targetUrl, "uid") {
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
		if s.Cookies != "" {
			req.Header.Set("Cookie", s.Cookies)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		targetUrl = resp.Request.URL.String()
	}

	// 2. Parse URL to get host suffix and uid/session
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(u.Host, ".")
	var hostSuffix string
	if len(parts) >= 3 {
		// live.shopee.sg -> sg
		hostSuffix = parts[len(parts)-1]
	} else {
		// fallback
		hostSuffix = "sg"
	}

	uid := u.Query().Get("uid")
	sessionId := u.Query().Get("session")

	apiHost := fmt.Sprintf("https://live.shopee.%s", hostSuffix)

	// 3. Call API
	if uid != "" {
		// Check ongoing live
		apiUrl := fmt.Sprintf("%s/api/v1/shop_page/live/ongoing?uid=%s", apiHost, uid)
		req, err := http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
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

		if data, ok := jsonResult["data"].(map[string]interface{}); ok {
			if ongoingLive, ok := data["ongoing_live"].(map[string]interface{}); ok {
				if sid, ok := ongoingLive["session_id"].(float64); ok {
					sessionId = fmt.Sprintf("%.0f", sid)
				} else if sid, ok := ongoingLive["session_id"].(string); ok {
					sessionId = sid
				}
			}
		}
	}

	if sessionId == "" {
		return nil, fmt.Errorf("session id not found")
	}

	// 4. Get Stream URL using Session ID
	apiUrl := fmt.Sprintf("%s/api/v1/session/%s", apiHost, sessionId)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
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

	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if session, ok := data["session"].(map[string]interface{}); ok {
			if playUrl, ok := session["play_url"].(string); ok && playUrl != "" {
				return &StreamInfo{Url: playUrl}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
