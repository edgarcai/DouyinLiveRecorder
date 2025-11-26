package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type TikTokSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewTikTokSpider() *TikTokSpider {
	return &TikTokSpider{
		Client: &http.Client{},
	}
}

func (t *TikTokSpider) SetProxy(proxyUrl string) {
	t.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			t.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (t *TikTokSpider) SetCookies(cookies string) {
	t.Cookies = cookies
}

func (t *TikTokSpider) GetStreamUrl(url string) (*StreamInfo, error) {
	// Headers
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://www.tiktok.com/")

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	htmlStr := string(bodyBytes)

	// Extract SIGI_STATE
	// <script id="SIGI_STATE" type="application/json">(.*?)</script>
	re := regexp.MustCompile(`<script id="SIGI_STATE" type="application/json">(.*?)</script>`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("SIGI_STATE not found in tiktok page")
	}

	jsonStr := matches[1]
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse SIGI_STATE: %v", err)
	}

	// Navigate JSON: LiveRoom -> liveRoomUserInfo -> liveRoom -> streamData -> pull_data -> stream_data
	liveRoom, ok := data["LiveRoom"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("LiveRoom not found")
	}

	liveRoomUserInfo, ok := liveRoom["liveRoomUserInfo"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("liveRoomUserInfo not found")
	}

	user, ok := liveRoomUserInfo["user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user info not found")
	}

	status, ok := user["status"].(float64)
	if !ok || status != 2 {
		return nil, fmt.Errorf("user is not live (status: %v)", status)
	}

	anchorName := ""
	if nickname, ok := user["nickname"].(string); ok {
		anchorName = nickname
	}

	innerLiveRoom, ok := liveRoom["liveRoom"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("inner liveRoom not found")
	}

	title := ""
	if t, ok := innerLiveRoom["title"].(string); ok {
		title = t
	}

	streamDataObj, ok := innerLiveRoom["streamData"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("streamData not found")
	}

	pullData, ok := streamDataObj["pull_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("pull_data not found")
	}

	streamDataStr, ok := pullData["stream_data"].(string)
	if !ok {
		return nil, fmt.Errorf("stream_data string not found")
	}

	var streamData map[string]interface{}
	if err := json.Unmarshal([]byte(streamDataStr), &streamData); err != nil {
		return nil, fmt.Errorf("failed to parse stream_data json: %v", err)
	}

	dataInner, ok := streamData["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("stream_data.data not found")
	}

	// Get FLV URL (sd-flv or similar)
	// data -> sd-flv -> main -> flv
	if sdFlv, ok := dataInner["sd-flv"].(map[string]interface{}); ok {
		if main, ok := sdFlv["main"].(map[string]interface{}); ok {
			if flv, ok := main["flv"].(string); ok {
				return &StreamInfo{
					Url:        flv,
					Title:      title,
					AnchorName: anchorName,
				}, nil
			}
		}
	}

	// Fallback loop
	for _, v := range dataInner {
		if vMap, ok := v.(map[string]interface{}); ok {
			if main, ok := vMap["main"].(map[string]interface{}); ok {
				if flv, ok := main["flv"].(string); ok {
					return &StreamInfo{
						Url:        flv,
						Title:      title,
						AnchorName: anchorName,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no stream url found in tiktok data")
}
