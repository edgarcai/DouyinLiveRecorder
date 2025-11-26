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

type YYSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewYYSpider() *YYSpider {
	return &YYSpider{}
}

func (s *YYSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *YYSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *YYSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Fetch HTML
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")
	req.Header.Set("Referer", "https://www.yy.com/")
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

	// 2. Extract cid and anchor name
	reCid := regexp.MustCompile(`sid : "(.*?)",\n\s+ssid`)
	matchesCid := reCid.FindStringSubmatch(htmlStr)
	if len(matchesCid) < 2 {
		return nil, fmt.Errorf("failed to find cid/sid")
	}
	cid := matchesCid[1]

	// 3. Call stream API
	// https://stream-manager.yy.com/v3/channel/streams
	params := url.Values{}
	params.Set("uid", "0")
	params.Set("cid", cid)
	params.Set("sid", cid)
	params.Set("appid", "0")
	params.Set("sequence", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Set("encode", "json")

	apiUrl := fmt.Sprintf("https://stream-manager.yy.com/v3/channel/streams?%s", params.Encode())

	data := fmt.Sprintf(`{"head":{"seq":%d,"appidstr":"0","bidstr":"121","cidstr":"%s","sidstr":"%s","uid64":0,"client_type":108,"client_ver":"5.17.0","stream_sys_ver":1,"app":"yylive_web","playersdk_ver":"5.17.0","thundersdk_ver":"0","streamsdk_ver":"5.17.0"},"client_attribute":{"client":"web","model":"web0","cpu":"","graphics_card":"","os":"chrome","osversion":"0","vsdk_version":"","app_identify":"","app_version":"","business":"","width":"1920","height":"1080","scale":"","client_type":8,"h265":0},"avp_parameter":{"version":1,"client_type":8,"service_type":0,"imsi":0,"send_time":%d,"line_seq":-1,"gear":4,"ssl":1,"stream_format":0}}`,
		time.Now().UnixMilli(), cid, cid, time.Now().Unix())

	reqApi, err := http.NewRequest("POST", apiUrl, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	reqApi.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")
	reqApi.Header.Set("Referer", "https://www.yy.com/")
	reqApi.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	respApi, err := client.Do(reqApi)
	if err != nil {
		return nil, err
	}
	defer respApi.Body.Close()

	apiBody, err := io.ReadAll(respApi.Body)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(apiBody, &jsonResult); err != nil {
		return nil, err
	}

	// 4. Extract stream URL
	if avpInfo, ok := jsonResult["avp_info_res"].(map[string]interface{}); ok {
		if streamLineAddr, ok := avpInfo["stream_line_addr"].(map[string]interface{}); ok {
			// Prefer CDN with highest quality/stability, usually first one or specific key
			// Logic from Python: usually iterates and picks one.
			// Here we try to find a valid HLS/FLV url.
			// The structure is complex, simplified extraction:
			for _, v := range streamLineAddr {
				if cdnInfo, ok := v.(map[string]interface{}); ok {
					if cdnUrl, ok := cdnInfo["cdn_info"].(map[string]interface{}); ok {
						if urlVal, ok := cdnUrl["url"].(string); ok {
							return &StreamInfo{Url: urlVal}, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
