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

type TaobaoSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewTaobaoSpider() *TaobaoSpider {
	return &TaobaoSpider{}
}

func (s *TaobaoSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *TaobaoSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *TaobaoSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Live ID
	// https://huodong.m.taobao.com/wow/zhibo/act/tpl-daily?id=123456
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}
	liveId := u.Query().Get("id")

	if liveId == "" {
		// Try fetching to get redirect
		client := &http.Client{}
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
		req.Header.Set("Referer", "https://huodong.m.taobao.com/")
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

		re := regexp.MustCompile(`var url = '(.*?)';`)
		matches := re.FindStringSubmatch(htmlStr)
		if len(matches) >= 2 {
			redirectUrl := matches[1]
			u2, err := url.Parse(redirectUrl)
			if err == nil {
				liveId = u2.Query().Get("id")
			}
		}
	}

	if liveId == "" {
		return nil, fmt.Errorf("failed to extract live id")
	}

	// 2. Call API
	// mtop.mediaplatform.live.livedetail
	params := url.Values{}
	params.Set("jsv", "2.7.0")
	params.Set("appKey", "12574478")
	params.Set("t", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Set("sign", "") // Sign seems optional or empty in Python code? Python code has 'sign': ''
	params.Set("AntiFlood", "true")
	params.Set("AntiCreep", "true")
	params.Set("api", "mtop.mediaplatform.live.livedetail")
	params.Set("v", "4.0")
	params.Set("preventFallback", "true")
	params.Set("type", "jsonp")
	params.Set("dataType", "jsonp")
	params.Set("callback", "mtopjsonp1")

	dataMap := map[string]string{
		"liveId": liveId,
	}
	dataJson, _ := json.Marshal(dataMap)
	params.Set("data", string(dataJson))

	apiUrl := fmt.Sprintf("https://h5api.m.taobao.com/h5/mtop.mediaplatform.live.livedetail/4.0/?%s", params.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Referer", "https://huodong.m.taobao.com/")
	// Cookie is critical for Taobao, especially _m_h5_tk
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
	bodyStr := string(bodyBytes)

	// 3. Parse JSONP
	// mtopjsonp1({...})
	start := strings.Index(bodyStr, "(")
	end := strings.LastIndex(bodyStr, ")")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("invalid jsonp response")
	}
	jsonStr := bodyStr[start+1 : end]

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonResult); err != nil {
		return nil, err
	}

	// 4. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if liveUrl, ok := data["liveUrl"].(string); ok && liveUrl != "" {
			return &StreamInfo{Url: liveUrl}, nil
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
