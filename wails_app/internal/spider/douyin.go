package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"wails_app/internal/pkg/crypto"
)

type DouyinSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewDouyinSpider() *DouyinSpider {
	return &DouyinSpider{
		Client: &http.Client{},
	}
}

func (d *DouyinSpider) SetProxy(proxyUrl string) {
	d.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			d.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (d *DouyinSpider) SetCookies(cookies string) {
	d.Cookies = cookies
}

func (d *DouyinSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// Extract web_rid
	// url format: https://live.douyin.com/123456
	parts := strings.Split(targetUrl, "live.douyin.com/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid douyin url")
	}
	webRid := strings.Split(parts[1], "?")[0]

	// Headers
	userAgent := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36 Core/1.116.567.400 QQBrowser/19.7.6764.400"
	// Use configured cookies if available, otherwise default
	cookie := d.Cookies
	if cookie == "" {
		cookie = "ttwid=1%7C2iDIYVmjzMcpZ20fcaFde0VghXAA3NaNXE_SLR68IyE%7C1761045455%7Cab35197d5cfb21df6cbb2fa7ef1c9262206b062c315b9d04da746d0b37dfbc7d"
	}

	// Params
	params := url.Values{}
	params.Set("aid", "6383")
	params.Set("app_name", "douyin_web")
	params.Set("live_id", "1")
	params.Set("device_platform", "web")
	params.Set("language", "zh-CN")
	params.Set("browser_language", "zh-CN")
	params.Set("browser_platform", "Win32")
	params.Set("browser_name", "Chrome")
	params.Set("browser_version", "116.0.0.0")
	params.Set("web_rid", webRid)
	params.Set("msToken", "")

	// Generate Signature
	query := params.Encode()
	aBogus := crypto.GenerateSignature(query, userAgent)
	apiUrl := fmt.Sprintf("https://live.douyin.com/webcast/room/web/enter/?%s&a_bogus=%s", query, aBogus)

	// Request
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", targetUrl)

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Navigate JSON to find stream URL
	// data -> data[0] -> stream_url -> flv_pull_url -> FULL_HD1
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response structure: data")
	}

	roomDataList, ok := data["data"].([]interface{})
	if !ok || len(roomDataList) == 0 {
		return nil, fmt.Errorf("room data not found or empty")
	}

	roomData := roomDataList[0].(map[string]interface{})
	status := roomData["status"].(float64)
	if status != 2 {
		return nil, fmt.Errorf("room is not live (status: %v)", status)
	}

	// Extract Metadata
	title := ""
	if t, ok := roomData["title"].(string); ok {
		title = t
	}

	anchorName := ""
	if owner, ok := roomData["owner"].(map[string]interface{}); ok {
		if nickname, ok := owner["nickname"].(string); ok {
			anchorName = nickname
		}
	}

	streamUrlObj, ok := roomData["stream_url"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stream_url not found")
	}

	// Try to get FLV pull url
	flvPullUrl, ok := streamUrlObj["flv_pull_url"].(map[string]interface{})
	if ok {
		// Just pick the first one for now, or "FULL_HD1"
		for _, v := range flvPullUrl {
			return &StreamInfo{
				Url:        v.(string),
				Title:      title,
				AnchorName: anchorName,
			}, nil
		}
	}

	// Fallback to HLS
	hlsPullUrlMap, ok := streamUrlObj["hls_pull_url_map"].(map[string]interface{})
	if ok {
		for _, v := range hlsPullUrlMap {
			return &StreamInfo{
				Url:        v.(string),
				Title:      title,
				AnchorName: anchorName,
			}, nil
		}
	}

	return nil, fmt.Errorf("no stream url found")
}

// Helper to extract regex
func extractRegex(content, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
