package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type KuaishouSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewKuaishouSpider() *KuaishouSpider {
	return &KuaishouSpider{
		Client: &http.Client{},
	}
}

func (k *KuaishouSpider) SetProxy(proxyUrl string) {
	k.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			k.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (k *KuaishouSpider) SetCookies(cookies string) {
	k.Cookies = cookies
}

func (k *KuaishouSpider) GetStreamUrl(url string) (*StreamInfo, error) {
	// Headers
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Set("Cookie", "did=web_e988652e11b545469633396abe85a89f; didv=1796004001000") // 基本 cookie

	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	htmlStr := string(bodyBytes)

	// 提取 window.__INITIAL_STATE__
	// <script>window.__INITIAL_STATE__=(.*?);\(function\(\)\{var s;
	re := regexp.MustCompile(`<script>window.__INITIAL_STATE__=(.*?);\(function\(\)\{var s;`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("stream info not found in kuaishou page")
	}

	jsonStr := matches[1]
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse stream info: %v", err)
	}

	// 导航 JSON: liveroom -> liveStream -> playUrls -> h264 -> adaptationSet -> representation
	// 注意：结构可能会根据 Python 代码分析而有所不同
	// Python 代码: play_list = re.findall('(\\{"liveStream".*?),"gameInfo', json_str)[0] + "}"

	liveStreamObj, ok := data["liveStream"].(map[string]interface{})
	if !ok {
		// 如果需要，尝试在嵌套结构中查找，或者正则表达式可能捕获了不同的级别
		return nil, fmt.Errorf("liveStream not found in INITIAL_STATE")
	}

	playUrls, ok := liveStreamObj["playUrls"].(map[string]interface{})
	if !ok || len(playUrls) == 0 { // 未找到或为空的组合条件
		return nil, fmt.Errorf("playUrls is empty")
	}

	h264, ok := playUrls["h264"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("livestream data not found")
	}

	adaptationSet, ok := h264["adaptationSet"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("adaptationSet not found")
	}

	representation, ok := adaptationSet["representation"].([]interface{})
	if !ok || len(representation) == 0 {
		return nil, fmt.Errorf("representation is empty")
	}

	// 迭代以找到最佳质量（通常是第一个或检查比特率）
	for _, rep := range representation {
		repMap, ok := rep.(map[string]interface{})
		if !ok {
			continue
		}
		urlStr, ok := repMap["url"].(string)
		if ok && urlStr != "" {
			return &StreamInfo{Url: urlStr}, nil
		}
	}

	return nil, fmt.Errorf("no valid stream found")
}
