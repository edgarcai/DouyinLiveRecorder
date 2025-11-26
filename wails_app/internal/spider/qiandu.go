package spider

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type QianduSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewQianduSpider() *QianduSpider {
	return &QianduSpider{}
}

func (s *QianduSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *QianduSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *QianduSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Fetch HTML
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Referer", "https://qiandurebo.com/web/index.php")
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
	htmlStr := string(bodyBytes)

	// 2. Extract Data
	// var user = { ... };
	re := regexp.MustCompile(`var user = (.*?)\r\n\s+user\.play_url`)
	matches := re.FindStringSubmatch(htmlStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find user data")
	}
	data := matches[1]

	// 3. Extract play_url
	rePlayUrl := regexp.MustCompile(`"play_url": "(.*?)",`)
	matchesPlayUrl := rePlayUrl.FindStringSubmatch(data)
	if len(matchesPlayUrl) >= 2 {
		playUrl := matchesPlayUrl[1]
		if playUrl != "" {
			return &StreamInfo{Url: playUrl}, nil
		}
	}

	return nil, fmt.Errorf("stream not found")
}
