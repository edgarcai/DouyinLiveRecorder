package spider

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

type YoutubeSpider struct {
	Client   *http.Client
	ProxyUrl string
	Cookies  string
}

func NewYoutubeSpider() *YoutubeSpider {
	return &YoutubeSpider{
		Client: &http.Client{},
	}
}

func (y *YoutubeSpider) SetProxy(proxyUrl string) {
	y.ProxyUrl = proxyUrl
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err == nil {
			y.Client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}
}

func (y *YoutubeSpider) SetCookies(cookies string) {
	y.Cookies = cookies
}

func (y *YoutubeSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// Use yt-dlp to get the stream URL
	// Ensure yt-dlp is installed
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found, please install it to record Youtube")
	}

	args := []string{"-g", targetUrl}
	if y.ProxyUrl != "" {
		args = append(args, "--proxy", y.ProxyUrl)
	}
	if y.Cookies != "" {
		// yt-dlp expects cookies file usually, but we can try passing via header or just rely on it handling it if passed as arg?
		// Actually passing raw cookies string to yt-dlp is tricky.
		// It supports --cookies-from-browser or --cookies file.
		// For now, let's ignore raw cookies string for yt-dlp unless we write to temp file.
		// Or use --add-header "Cookie: ..."
		args = append(args, "--add-header", fmt.Sprintf("Cookie:%s", y.Cookies))
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

	// yt-dlp might return two lines (video and audio) for some formats.
	// We want the best combined or just the first one.
	lines := strings.Split(streamUrl, "\n")
	if len(lines) > 0 {
		return &StreamInfo{Url: lines[0]}, nil
	}

	return &StreamInfo{Url: streamUrl}, nil
}
