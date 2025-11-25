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

func (k *KuaishouSpider) GetStreamUrl(url string) (string, error) {
	// Headers
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Set("Cookie", "did=web_e988652e11b545469633396abe85a89f; didv=1796004001000") // Basic cookie

	resp, err := k.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	htmlStr := string(bodyBytes)

	// Extract window.__INITIAL_STATE__
	// <script>window.__INITIAL_STATE__=(.*?);\(function\(\)\{var s;
	re := regexp.MustCompile(`<script>window.__INITIAL_STATE__=(.*?);\(function\(\)\{var s;`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return "", fmt.Errorf("INITIAL_STATE not found in kuaishou page")
	}

	jsonStr := matches[1]
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("failed to parse INITIAL_STATE: %v", err)
	}

	// Navigate JSON: liveroom -> liveStream -> playUrls -> h264 -> adaptationSet -> representation
	// Note: The structure might vary based on the Python code analysis
	// Python code: play_list = re.findall('(\\{"liveStream".*?),"gameInfo', json_str)[0] + "}"
	// It seems the JSON structure in INITIAL_STATE is huge, and we might need to look for liveStream directly if the full parse fails or is too complex.
	// But let's try to navigate the parsed map first.

	liveStreamObj, ok := data["liveStream"].(map[string]interface{})
	if !ok {
		// Try to find it in a nested structure if needed, or maybe the regex captured a different level
		return "", fmt.Errorf("liveStream not found in INITIAL_STATE")
	}

	playUrls, ok := liveStreamObj["playUrls"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("playUrls not found")
	}

	h264, ok := playUrls["h264"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("h264 playUrls not found")
	}

	adaptationSet, ok := h264["adaptationSet"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("adaptationSet not found")
	}

	representation, ok := adaptationSet["representation"].([]interface{})
	if !ok || len(representation) == 0 {
		return "", fmt.Errorf("representation list not found or empty")
	}

	// Iterate to find the best quality (usually the first one or check bitrate)
	for _, rep := range representation {
		repMap, ok := rep.(map[string]interface{})
		if !ok {
			continue
		}
		url, ok := repMap["url"].(string)
		if ok && url != "" {
			return url, nil
		}
	}

	return "", fmt.Errorf("no valid stream url found in representation")
}
