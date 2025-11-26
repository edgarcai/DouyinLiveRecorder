package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HuajiaoSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewHuajiaoSpider 创建 Huajiao 爬虫实例
func NewHuajiaoSpider() *HuajiaoSpider {
	return &HuajiaoSpider{}
}

// SetProxy 设置代理地址
func (s *HuajiaoSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *HuajiaoSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *HuajiaoSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://www.huajiao.com/l/123456
	parts := strings.Split(targetUrl, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Get Feed Info
	api := fmt.Sprintf("https://live.huajiao.com/feed/getFeedInfo?relateid=%s", roomId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "living/9.4.0 (com.huajiao.seeding; build:2410231746; iOS 17.0.0) Alamofire/9.4.0")
	req.Header.Set("Accept-Language", "zh-Hans-US;q=1.0")
	req.Header.Set("sdk_version", "1")
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

	if errmsg, ok := jsonResult["errmsg"].(string); ok && errmsg != "" {
		return nil, fmt.Errorf("huajiao error: %s", errmsg)
	}

	data, ok := jsonResult["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get data")
	}

	feed, ok := data["feed"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get feed")
	}
	author, ok := data["author"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get author")
	}

	sn, _ := feed["sn"].(string)
	liveid, _ := feed["relateid"].(string) // relateid is string in JSON usually
	if liveid == "" {
		// Try to cast if it's a number
		if lid, ok := feed["relateid"].(float64); ok {
			liveid = fmt.Sprintf("%.0f", lid)
		}
	}
	uid, _ := author["uid"].(string)
	if uid == "" {
		if u, ok := author["uid"].(float64); ok {
			uid = fmt.Sprintf("%.0f", u)
		}
	}

	// 3. Get Substream
	params := url.Values{}
	params.Set("time", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Set("version", "1.0.0")
	params.Set("sn", sn)
	params.Set("liveid", liveid)
	params.Set("uid", uid)
	params.Set("encode", "h265") // Or h264

	subApi := fmt.Sprintf("https://live.huajiao.com/live/substream?%s", params.Encode())
	reqSub, err := http.NewRequest("GET", subApi, nil)
	if err != nil {
		return nil, err
	}
	reqSub.Header.Set("User-Agent", "living/9.4.0 (com.huajiao.seeding; build:2410231746; iOS 17.0.0) Alamofire/9.4.0")
	reqSub.Header.Set("Accept-Language", "zh-Hans-US;q=1.0")
	reqSub.Header.Set("sdk_version", "1")
	if s.Cookies != "" {
		reqSub.Header.Set("Cookie", s.Cookies)
	}

	respSub, err := client.Do(reqSub)
	if err != nil {
		return nil, err
	}
	defer respSub.Body.Close()

	subBody, err := io.ReadAll(respSub.Body)
	if err != nil {
		return nil, err
	}

	var jsonSub map[string]interface{}
	if err := json.Unmarshal(subBody, &jsonSub); err != nil {
		return nil, err
	}

	if subData, ok := jsonSub["data"].(map[string]interface{}); ok {
		if h264Url, ok := subData["h264_url"].(string); ok && h264Url != "" {
			return &StreamInfo{Url: h264Url}, nil
		}
		if mainUrl, ok := subData["main"].(string); ok && mainUrl != "" {
			return &StreamInfo{Url: mainUrl}, nil
		}
	}

	return nil, fmt.Errorf("stream not found")
}
