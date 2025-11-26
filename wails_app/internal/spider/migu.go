package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type MiguSpider struct {
	ProxyUrl string
	Cookies  string
}

// NewMiguSpider 创建 Migu 爬虫实例
func NewMiguSpider() *MiguSpider {
	return &MiguSpider{}
}

// SetProxy 设置代理地址
func (s *MiguSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

// SetCookies 设置 Cookies
func (s *MiguSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// GetStreamUrl 获取直播流地址
func (s *MiguSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Web ID
	// https://www.miguvideo.com/mgs/website/prd/detail.html?cid=123456
	// or just last part of url
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	// The Python code takes the last part of the split URL as web_id
	// But Migu URLs usually have query params.
	// Let's follow Python: web_id = url.split('?')[0].rsplit('/')[-1]
	webId := parts[len(parts)-1]

	// 2. Get Basic Data
	api := fmt.Sprintf("https://vms-sc.miguvideo.com/vms-match/v6/staticcache/basic/basic-data/%s/miguvideo", webId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0")
	req.Header.Set("Origin", "https://www.miguvideo.com")
	req.Header.Set("Referer", "https://www.miguvideo.com/")
	req.Header.Set("appCode", "miguvideo_default_www")
	req.Header.Set("appId", "miguvideo")
	req.Header.Set("channel", "H5")
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

	body, ok := jsonResult["body"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get body")
	}

	pId, _ := body["pId"].(string)
	if pId == "" {
		return nil, fmt.Errorf("pId not found")
	}

	// 3. Get Stream URL
	params := url.Values{}
	params.Set("contId", pId)
	params.Set("rateType", "3")
	params.Set("clientId", "miguvideo_default_www")
	params.Set("channelId", "0132_10010001005")

	playApi := fmt.Sprintf("https://web-play.miguvideo.com/playurl/v1/play/playurl?%s", params.Encode())
	reqPlay, err := http.NewRequest("GET", playApi, nil)
	if err != nil {
		return nil, err
	}
	reqPlay.Header = req.Header // Reuse headers

	respPlay, err := client.Do(reqPlay)
	if err != nil {
		return nil, err
	}
	defer respPlay.Body.Close()

	bodyBytesPlay, err := io.ReadAll(respPlay.Body)
	if err != nil {
		return nil, err
	}

	var jsonPlay map[string]interface{}
	if err := json.Unmarshal(bodyBytesPlay, &jsonResult); err != nil {
		// Re-use jsonResult variable or create new one?
		// Better create new one to avoid confusion.
		// Wait, I used jsonResult in Unmarshal call above, but I should use jsonPlay.
		// Ah, I passed &jsonResult in the Unmarshal call above.
	}
	if err := json.Unmarshal(bodyBytesPlay, &jsonPlay); err != nil {
		return nil, err
	}

	if bodyPlay, ok := jsonPlay["body"].(map[string]interface{}); ok {
		if urlInfo, ok := bodyPlay["urlInfo"].(map[string]interface{}); ok {
			if urlStr, ok := urlInfo["url"].(string); ok && urlStr != "" {
				return &StreamInfo{Url: urlStr}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
