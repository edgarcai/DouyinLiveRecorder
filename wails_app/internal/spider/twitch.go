package spider

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

type TwitchSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewTwitchSpider() *TwitchSpider {
	return &TwitchSpider{
		Client: &http.Client{},
	}
}

func (t *TwitchSpider) SetProxy(proxyUrl string) {
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

func (t *TwitchSpider) SetCookies(cookies string) {
	t.Cookies = cookies
}

func (t *TwitchSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// Use yt-dlp for Twitch as well, it's reliable
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found, please install it to record Twitch")
	}

	args := []string{"-g", targetUrl}
	if t.ProxyUrl != "" {
		args = append(args, "--proxy", t.ProxyUrl)
	}
	if t.Cookies != "" {
		args = append(args, "--add-header", fmt.Sprintf("Cookie:%s", t.Cookies))
	}

	cmd := exec.Command("yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %v", err)
	}

	streamUrl := strings.TrimSpace(string(out))
	if streamUrl == "" {
		return nil, fmt.Errorf("no stream url found by yt-dlp")
	}

	lines := strings.Split(streamUrl, "\n")
	if len(lines) > 0 {
		return &StreamInfo{Url: lines[0]}, nil
	}

	return &StreamInfo{Url: streamUrl}, nil
}
