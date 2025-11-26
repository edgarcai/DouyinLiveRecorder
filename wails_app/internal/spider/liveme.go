package spider

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LiveMeSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewLiveMeSpider 创建 LiveMe 爬虫实例
func NewLiveMeSpider() *LiveMeSpider {
	return &LiveMeSpider{}
}

// SetProxy 设置代理地址
func (s *LiveMeSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *LiveMeSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

const (
	liveMeAm = "LM6000101139961122666757"
	liveMeRl = "undefined"
)

// createRandom 生成指定长度的随机字符串
func (s *LiveMeSpider) createRandom(length int) string {
	const characters = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
	b := make([]byte, length)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

// createSignature 生成签名
func (s *LiveMeSpider) createSignature(input string) string {
	if input == "" {
		input = "4l4m5"
	}
	signature := ""
	number := 0
	for _, charCode := range input {
		if charCode >= '0' && charCode <= '9' {
			number = number*10 + int(charCode-'0')
		} else {
			if number != 0 {
				signature += s.createRandom(number)
				number = 0
			}
			signature += string(charCode)
		}
	}
	if number != 0 {
		signature += s.createRandom(number)
	}
	return signature
}

// md5Hash 计算 MD5 哈希
func (s *LiveMeSpider) md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// requestSign 生成请求签名
func (s *LiveMeSpider) requestSign(signParams map[string]interface{}) string {
	// lm_s_key = atob('ZGQ0NmRiYjQ0MmI2ZTRiYTgxN2Q2MzQ3ZDJkZGY0OTM=');
	lmSKeyBytes, _ := base64.StdEncoding.DecodeString("ZGQ0NmRiYjQ0MmI2ZTRiYTgxN2Q2MzQ3ZDJkZGY0OTM=")
	lmSKey := string(lmSKeyBytes)

	// Sort keys
	keys := make([]string, 0, len(signParams))
	for k := range signParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sKeyBuilder strings.Builder
	for _, k := range keys {
		val := signParams[k]
		valStr := ""
		switch v := val.(type) {
		case string:
			valStr = v
		case int:
			valStr = strconv.Itoa(v)
		case int64:
			valStr = strconv.FormatInt(v, 10)
		case float64:
			valStr = fmt.Sprintf("%.0f", v) // Assuming integers for now based on JS logic
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		sKeyBuilder.WriteString(k)
		sKeyBuilder.WriteString(valStr)
	}

	lmSId, _ := signParams["lm_s_id"].(string)
	lmSTs, _ := signParams["lm_s_ts"].(string)

	sKeyBuilder.WriteString(lmSId)
	sKeyBuilder.WriteString(lmSTs)
	sKeyBuilder.WriteString(lmSKey)

	return s.md5Hash(sKeyBuilder.String())
}

// GetStreamUrl 获取直播流地址
func (s *LiveMeSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://www.liveme.com/us/v/123456
	// or https://www.liveme.com/us/v/123456/index.html

	// If not direct video link, fetch to get og:url
	if !strings.Contains(targetUrl, "index.html") {
		client := &http.Client{}
		req, err := http.NewRequest("GET", targetUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
		if s.Cookies != "" {
			req.Header.Set("Cookie", s.Cookies)
		}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			htmlStr := string(bodyBytes)
			re := regexp.MustCompile(`<meta property="og:url" content="(.*?)">`)
			matches := re.FindStringSubmatch(htmlStr)
			if len(matches) >= 2 {
				targetUrl = matches[1]
			}
		}
	}

	parts := strings.Split(strings.Split(targetUrl, "/index.html")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Generate Sign Params
	ts := fmt.Sprintf("%d1", time.Now().UnixMilli()) // JS: new Date().getTime() + id(1) -> concatenation? No, `${new Date().getTime()}${id}`

	// s = Rm(r) where r is ts
	lmSStr := s.md5Hash(ts)

	vali := s.createSignature("4l4m5")

	dataI := map[string]interface{}{
		"lm_s_id":      liveMeAm,
		"lm_s_ts":      ts,
		"lm_s_str":     lmSStr,
		"lm_s_ver":     1,
		"h5":           1,
		"_time":        time.Now().UnixMilli(),
		"thirdchannel": 6,
		"videoid":      roomId,
		"area":         "zh",
		"vali":         vali,
	}

	signParams := map[string]interface{}{
		"alias":             "liveme",
		"tongdun_black_box": "",
		"os":                "web",
	}
	for k, v := range dataI {
		signParams[k] = v
	}

	lmSSign := s.requestSign(signParams)
	signParams["lm_s_sign"] = lmSSign

	// 3. Call API
	params := url.Values{}
	params.Set("alias", "liveme")
	params.Set("tongdun_black_box", "")
	params.Set("os", "web")

	api := fmt.Sprintf("https://live.liveme.com/live/queryinfosimple?%s", params.Encode())

	// Prepare POST data
	postData := url.Values{}
	for k, v := range signParams {
		valStr := ""
		switch val := v.(type) {
		case string:
			valStr = val
		case int:
			valStr = strconv.Itoa(val)
		case int64:
			valStr = strconv.FormatInt(val, 10)
		case float64:
			valStr = fmt.Sprintf("%.0f", val)
		}
		postData.Set(k, valStr)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", api, strings.NewReader(postData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("Origin", "https://www.liveme.com")
	req.Header.Set("Referer", "https://www.liveme.com")
	req.Header.Set("lm-s-sign", lmSSign)
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

	// 4. Extract Stream URL
	if data, ok := jsonResult["data"].(map[string]interface{}); ok {
		if videoInfo, ok := data["video_info"].(map[string]interface{}); ok {
			if status, ok := videoInfo["status"].(string); ok && status == "0" {
				if m3u8, ok := videoInfo["hlsvideosource"].(string); ok && m3u8 != "" {
					return &StreamInfo{Url: m3u8}, nil
				}
				if flv, ok := videoInfo["videosource"].(string); ok && flv != "" {
					return &StreamInfo{Url: flv}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
