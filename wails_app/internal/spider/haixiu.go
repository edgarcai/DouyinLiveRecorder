package spider

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type HaixiuSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewHaixiuSpider 创建 Haixiu 爬虫实例
func NewHaixiuSpider() *HaixiuSpider {
	return &HaixiuSpider{}
}

// SetProxy 设置代理地址
func (s *HaixiuSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *HaixiuSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// md5Hash 计算 MD5 哈希
func (s *HaixiuSpider) md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// sign 生成签名
func (s *HaixiuSpider) sign(params map[string]string) string {
	// Secret key derived from haixiu.js analysis
	// _a123="haija1c7", _b2x="xiuhc2a6", _c3y="anchc3a5", _dx34="famic7a2"
	// substring(4) -> "a1c7" + "c2a6" + "c3a5" + "c7a2"
	const secretKey = "a1c7c2a6c3a5c7a2"

	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
		if i < len(keys)-1 {
			sb.WriteString("&")
		}
	}

	sb.WriteString(secretKey)
	return s.md5Hash(sb.String())
}

// GetStreamUrl 获取直播流地址
func (s *HaixiuSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://www.haixiutv.com/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	isHaixiu := strings.Contains(targetUrl, "haixiutv")
	var accessToken string
	if isHaixiu {
		accessToken = "pLXSC%2FXJ0asc1I21tVL5FYZhNJn2Zg6d7m94umCnpgL%2BuVm31GQvyw%3D%3D"
	} else {
		accessToken = "s7FUbTJ%2BjILrR7kicJUg8qr025ZVjd07DAnUQd8c7g%2Fo4OH9pdSX6w%3D%3D"
	}

	// 2. Prepare Params
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())

	params := map[string]string{
		"accessToken": accessToken,
		"tku":         "3000006",
		"c":           "10138100100000",
		"_st1":        timestamp,
	}

	// 3. Generate Sign
	ajaxData := s.sign(params)

	// 4. Update Params for API Call
	decodedToken, _ := url.QueryUnescape(accessToken)
	decodedToken, _ = url.QueryUnescape(decodedToken) // Double unquote in Python code

	// Wait, Python code:
	// params["accessToken"] = urllib.parse.unquote(urllib.parse.unquote(access_token))
	// params['_ajaxData1'] = ajax_data
	// params['_'] = int(time.time() * 1000)

	apiParams := url.Values{}
	apiParams.Set("accessToken", decodedToken)
	apiParams.Set("tku", "3000006")
	apiParams.Set("c", "10138100100000")
	apiParams.Set("_st1", timestamp)
	apiParams.Set("_ajaxData1", ajaxData)
	apiParams.Set("_", timestamp)

	var api string
	headers := make(http.Header)
	headers.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	if s.Cookies != "" {
		headers.Set("Cookie", s.Cookies)
	}

	if isHaixiu {
		api = fmt.Sprintf("https://service.haixiutv.com/v2/room/%s/media/advanceInfoRoom?%s", roomId, apiParams.Encode())
		headers.Set("Origin", "https://www.haixiutv.com")
		headers.Set("Referer", "https://www.haixiutv.com/")
	} else {
		api = fmt.Sprintf("https://service.lehaitv.com/v2/room/%s/media/advanceInfoRoom?%s", roomId, apiParams.Encode())
		headers.Set("Origin", "https://www.lehaitv.com")
		headers.Set("Referer", "https://www.lehaitv.com")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers

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

	// 5. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if liveStatus, ok := data["liveStatus"].(bool); ok && liveStatus {
			if urlStr, ok := data["url"].(string); ok && urlStr != "" {
				return &StreamInfo{Url: urlStr}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
