package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type BaiduSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewBaiduSpider() *BaiduSpider {
	return &BaiduSpider{}
}

func (s *BaiduSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *BaiduSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *BaiduSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	re := regexp.MustCompile(`room_id=(.*?)&`)
	matches := re.FindStringSubmatch(targetUrl)
	if len(matches) < 2 {
		re2 := regexp.MustCompile(`room_id=(.*)`)
		matches = re2.FindStringSubmatch(targetUrl)
		if len(matches) < 2 {
			return nil, fmt.Errorf("failed to extract room id")
		}
	}
	roomId := matches[1]

	// 2. Prepare API Call
	uids := []string{
		"h5-683e85bdf741bf2492586f7ca39bf465",
		"h5-c7c6dc14064a136be4215b452fab9eea",
		"h5-4581281f80bb8968bd9a9dfba6050d3a",
	}
	uid := uids[rand.Intn(len(uids))]

	params := url.Values{}
	params.Set("cmd", "371")
	params.Set("action", "star")
	params.Set("service", "bdbox")
	params.Set("osname", "baiduboxapp")

	dataJson := fmt.Sprintf(`{"data":{"room_id":"%s","device_id":"h5-683e85bdf741bf2492586f7ca39bf465","source_type":0,"osname":"baiduboxapp"},"replay_slice":0,"nid":"","schemeParams":{"src_pre":"pc","src_suf":"other","bd_vid":"","share_uid":"","share_cuk":"","share_ecid":"","zb_tag":"","shareTaskInfo":"{\"room_id\":\"%s\"}","share_from":"","ext_params":"","nid":""}}`, roomId, roomId)
	params.Set("data", dataJson)

	params.Set("ua", "360_740_ANDROID_0")
	params.Set("bd_vid", "")
	params.Set("uid", uid)
	params.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))

	apiUrl := fmt.Sprintf("https://mbd.baidu.com/searchbox?%s", params.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Referer", "https://live.baidu.com/")
	req.Header.Set("Connection", "keep-alive")
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

	// 3. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		// The key is dynamic, usually the first one
		for _, v := range data {
			if roomData, ok := v.(map[string]interface{}); ok {
				if status, ok := roomData["status"].(string); ok && status == "0" {
					if video, ok := roomData["video"].(map[string]interface{}); ok {
						prefix := "https://hls.liveshow.bdstatic.com/live/"

						// Try url_clarity_list
						if clarityList, ok := video["url_clarity_list"].([]interface{}); ok && len(clarityList) > 0 {
							for _, item := range clarityList {
								if itemMap, ok := item.(map[string]interface{}); ok {
									if urls, ok := itemMap["urls"].(map[string]interface{}); ok {
										if flv, ok := urls["flv"].(string); ok {
											// Construct m3u8 from flv path
											// prefix + flv.rsplit('.', 1)[0].rsplit('/', 1)[1] + '.m3u8'
											parts := strings.Split(flv, "/")
											if len(parts) > 0 {
												filename := parts[len(parts)-1]
												filename = strings.TrimSuffix(filename, ".flv")
												return &StreamInfo{Url: prefix + filename + ".m3u8"}, nil
											}
										}
									}
								}
							}
						}

						// Fallback to url_list
						if urlList, ok := video["url_list"].([]interface{}); ok {
							for _, item := range urlList {
								if itemMap, ok := item.(map[string]interface{}); ok {
									if urls, ok := itemMap["urls"].([]interface{}); ok && len(urls) > 0 {
										if urlObj, ok := urls[0].(map[string]interface{}); ok {
											if hls, ok := urlObj["hls"].(string); ok {
												// prefix + hls.rsplit('?', 1)[0].rsplit('/', 1)[1]
												parts := strings.Split(strings.Split(hls, "?")[0], "/")
												if len(parts) > 0 {
													return &StreamInfo{Url: prefix + parts[len(parts)-1]}, nil
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
			break // Only check the first key
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
