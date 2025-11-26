package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BilibiliSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewBilibiliSpider() *BilibiliSpider {
	return &BilibiliSpider{
		Client: &http.Client{},
	}
}

func (b *BilibiliSpider) SetProxy(proxyUrl string) {
	b.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			b.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (b *BilibiliSpider) SetCookies(cookies string) {
	b.Cookies = cookies
}

func (b *BilibiliSpider) GetStreamUrl(url string) (*StreamInfo, error) {
	// Logic ported from get_bilibili_stream_data
	// 1. Extract Room ID
	// url: https://live.bilibili.com/123
	parts := strings.Split(url, "live.bilibili.com/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid bilibili url")
	}
	roomId := strings.Split(parts[1], "?")[0]

	// 2. Get Room Info (to get real room id if it's a short id)
	infoUrl := fmt.Sprintf("https://api.live.bilibili.com/room/v1/Room/room_init?id=%s", roomId)
	resp, err := b.Client.Get(infoUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var infoData map[string]interface{}
	if err := json.Unmarshal(body, &infoData); err != nil {
		return nil, err
	}

	data, ok := infoData["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data not found")
	}

	liveStatus, _ := data["live_status"].(float64)
	if liveStatus != 1 {
		return nil, fmt.Errorf("room is not live")
	}

	realRoomId := fmt.Sprintf("%.0f", data["room_id"].(float64))

	// 3. Get Play Url
	// https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo
	playUrlApi := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo?room_id=%s&protocol=0,1&format=0,1,2&codec=0,1&qn=10000&platform=web&ptype=16", realRoomId)

	req, _ := http.NewRequest("GET", playUrlApi, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	resp2, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)

	var playData map[string]interface{}
	if err := json.Unmarshal(body2, &playData); err != nil {
		return nil, err
	}

	data2, ok := playData["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("playurl not found")
	}

	playUrlInfo, ok := data2["playurl_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("playurl_info not found")
	}

	playurl, ok := playUrlInfo["playurl"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("playurl object not found")
	}

	stream, ok := playurl["stream"].([]interface{})
	if !ok || len(stream) == 0 {
		return nil, fmt.Errorf("stream not found or empty")
	}

	// Get first stream, first format, first codec
	stream0 := stream[0].(map[string]interface{})
	format, _ := stream0["format"].([]interface{})
	format0 := format[0].(map[string]interface{})
	codec, _ := format0["codec"].([]interface{})
	codec0 := codec[0].(map[string]interface{})

	baseUrl := codec0["base_url"].(string)
	urlInfo, _ := codec0["url_info"].([]interface{})
	urlInfo0 := urlInfo[0].(map[string]interface{})
	host := urlInfo0["host"].(string)
	extra := urlInfo0["extra"].(string)

	// Final URL
	finalUrl := host + baseUrl + extra
	return &StreamInfo{Url: finalUrl}, nil
}
