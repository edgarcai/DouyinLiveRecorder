package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type KugouSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewKugouSpider() *KugouSpider {
	return &KugouSpider{}
}

func (s *KugouSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *KugouSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *KugouSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	var roomId string
	if strings.Contains(targetUrl, "roomId=") {
		re := regexp.MustCompile(`roomId=(\d+)`)
		matches := re.FindStringSubmatch(targetUrl)
		if len(matches) >= 2 {
			roomId = matches[1]
		}
	} else {
		parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
		if len(parts) > 0 {
			roomId = parts[len(parts)-1]
		}
	}

	if roomId == "" {
		return nil, fmt.Errorf("failed to extract room id")
	}

	// 2. Check Live Status
	apiUrl := fmt.Sprintf("https://service2.fanxing.kugou.com/roomcen/room/web/cdn/getEnterRoomInfo?roomId=%s", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://fanxing2.kugou.com/")
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
		if liveType, ok := data["liveType"].(float64); ok && liveType != -1 {
			// 3. Get Stream URL
			params := url.Values{}
			params.Set("std_rid", roomId)
			params.Set("std_plat", "7")
			params.Set("std_kid", "0")
			params.Set("streamType", "1-2-4-5-8")
			params.Set("ua", "fx-flash")
			params.Set("targetLiveTypes", "1-5-6")
			params.Set("version", "1000")
			params.Set("supportEncryptMode", "1")
			params.Set("appid", "1010")
			params.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))

			streamApi := fmt.Sprintf("https://fx1.service.kugou.com/video/pc/live/pull/mutiline/streamaddr?%s", params.Encode())
			reqStream, err := http.NewRequest("GET", streamApi, nil)
			if err != nil {
				return nil, err
			}
			reqStream.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
			reqStream.Header.Set("Referer", "https://fanxing2.kugou.com/")
			if s.Cookies != "" {
				reqStream.Header.Set("Cookie", s.Cookies)
			}

			respStream, err := client.Do(reqStream)
			if err != nil {
				return nil, err
			}
			defer respStream.Body.Close()

			streamBody, err := io.ReadAll(respStream.Body)
			if err != nil {
				return nil, err
			}

			var jsonStream map[string]interface{}
			if err := json.Unmarshal(streamBody, &jsonStream); err != nil {
				return nil, err
			}

			if dataStream, ok := jsonStream["data"].(map[string]interface{}); ok {
				if lines, ok := dataStream["lines"].([]interface{}); ok && len(lines) > 0 {
					// Use last line
					lastLine := lines[len(lines)-1]
					if lineMap, ok := lastLine.(map[string]interface{}); ok {
						if profiles, ok := lineMap["streamProfiles"].([]interface{}); ok && len(profiles) > 0 {
							if profile, ok := profiles[0].(map[string]interface{}); ok {
								if httpsFlv, ok := profile["httpsFlv"].([]interface{}); ok && len(httpsFlv) > 0 {
									if flv, ok := httpsFlv[0].(string); ok {
										return &StreamInfo{Url: flv}, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
