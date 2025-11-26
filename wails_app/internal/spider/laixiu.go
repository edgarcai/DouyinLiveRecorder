package spider

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LaixiuSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewLaixiuSpider 创建 Laixiu 爬虫实例
func NewLaixiuSpider() *LaixiuSpider {
	return &LaixiuSpider{}
}

// SetProxy 设置代理地址
func (s *LaixiuSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *LaixiuSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// md5Hash 计算 MD5 哈希
func (s *LaixiuSpider) md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// calculateSign 计算签名
func (s *LaixiuSpider) calculateSign() (int64, string, string) {
	a := time.Now().UnixMilli()
	sUuid := strings.ReplaceAll(uuid.New().String(), "-", "")
	u := "kk792f28d6ff1f34ec702c08626d454b39pro"

	inputStr := fmt.Sprintf("web%s%d%s", sUuid, a, u)
	md5Hash := s.md5Hash(inputStr)

	return a, sUuid, md5Hash
}

// GetStreamUrl 获取直播流地址
func (s *LaixiuSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	// https://www.laixiu.com/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	roomId := parts[len(parts)-1]

	// 2. Generate Sign
	timestamp, imei, requestId := s.calculateSign()

	// 3. Call API
	params := url.Values{}
	params.Set("c", "10138001000000")
	params.Set("tku", "3000006")
	params.Set("accessToken", "")
	params.Set("imei", imei)
	params.Set("timestamp", fmt.Sprintf("%d", timestamp))
	params.Set("requestId", requestId)

	api := fmt.Sprintf("https://service.laixiu.com/v2/room/%s/media/advanceInfoRoom?%s", roomId, params.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")
	req.Header.Set("Origin", "https://www.laixiu.com")
	req.Header.Set("Referer", "https://www.laixiu.com/")
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
		if liveStatus, ok := data["liveStatus"].(bool); ok && liveStatus {
			if urlStr, ok := data["url"].(string); ok && urlStr != "" {
				return &StreamInfo{Url: urlStr}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
