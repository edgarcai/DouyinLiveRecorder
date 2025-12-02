package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type WeiboSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewWeiboSpider() *WeiboSpider {
	return &WeiboSpider{}
}

func (s *WeiboSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *WeiboSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *WeiboSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	var roomId string

	// 1. 提取房间 ID
	if strings.Contains(targetUrl, "show/") {
		// https://weibo.com/l/wblive/p/show/1022:2321324990000000000000
		parts := strings.Split(strings.Split(targetUrl, "?")[0], "show/")
		if len(parts) > 1 {
			roomId = parts[1]
		}
	} else if strings.Contains(targetUrl, "/u/") {
		// https://weibo.com/u/5885340893
		parts := strings.Split(strings.Split(targetUrl, "?")[0], "/u/")
		if len(parts) > 1 {
			uid := parts[1]
			// 调用 mymblog API 查找直播状态
			apiUrl := fmt.Sprintf("https://weibo.com/ajax/statuses/mymblog?uid=%s&page=1&feature=0", uid)
			client := &http.Client{}
			req, err := http.NewRequest("GET", apiUrl, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
			req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
			req.Header.Set("Referer", targetUrl)
			if s.Cookies != "" {
				req.Header.Set("Cookie", s.Cookies)
			} else {
				// 来自 Python 脚本的默认 cookie
				req.Header.Set("Cookie", "XSRF-TOKEN=qAP-pIY5V4tO6blNOhA4IIOD; SUB=_2AkMRNMCwf8NxqwFRmfwWymPrbI9-zgzEieKnaDFrJRMxHRl-yT9kqmkhtRB6OrTuX5z9N_7qk9C3xxEmNR-8WLcyo2PM;")
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
				if list, ok := data["list"].([]interface{}); ok {
					for _, item := range list {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if pageInfo, ok := itemMap["page_info"].(map[string]interface{}); ok {
								if objType, ok := pageInfo["object_type"].(string); ok && objType == "live" {
									if objId, ok := pageInfo["object_id"].(string); ok {
										roomId = objId
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if roomId == "" {
		return nil, fmt.Errorf("failed to extract room id")
	}

	// 2. 调用直播 API
	apiUrl := fmt.Sprintf("https://weibo.com/l/pc/anchor/live?live_id=%s", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Referer", "https://weibo.com/")
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

	// 3. 提取流 URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if item, ok := data["item"].(map[string]interface{}); ok {
			if status, ok := item["status"].(float64); ok && status == 1 {
				if streamInfo, ok := item["stream_info"].(map[string]interface{}); ok {
					if pull, ok := streamInfo["pull"].(map[string]interface{}); ok {
						if m3u8, ok := pull["live_origin_hls_url"].(string); ok && m3u8 != "" {
							return &StreamInfo{Url: m3u8}, nil
						}
						if flv, ok := pull["live_origin_flv_url"].(string); ok && flv != "" {
							return &StreamInfo{Url: flv}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
