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
	// 使用 yt-dlp 获取流 URL
	// 确保已安装 yt-dlp
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found, please install it to record Youtube")
	}

	args := []string{"-g", targetUrl}
	if y.ProxyUrl != "" {
		args = append(args, "--proxy", y.ProxyUrl)
	}
	if y.Cookies != "" {
		// yt-dlp 通常期望 cookies 文件，但我们可以尝试通过 header 传递，或者如果作为参数传递，它是否能处理？
		// 实际上，将原始 cookies 字符串传递给 yt-dlp 很棘手。
		// 它支持 --cookies-from-browser 或 --cookies 文件。
		// 目前，除非我们写入临时文件，否则忽略 yt-dlp 的原始 cookies 字符串。
		// 或者使用 --add-header "Cookie: ..."
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

	// 对于某些格式，yt-dlp 可能会返回两行（视频和音频）。
	// 我们想要最佳组合或仅第一行。
	lines := strings.Split(streamUrl, "\n")
	if len(lines) > 0 {
		return &StreamInfo{Url: lines[0]}, nil
	}

	return &StreamInfo{Url: streamUrl}, nil
}
