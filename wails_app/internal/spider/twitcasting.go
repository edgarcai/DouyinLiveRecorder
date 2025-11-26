package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type TwitCastingSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewTwitCastingSpider() *TwitCastingSpider {
	return &TwitCastingSpider{}
}

func (s *TwitCastingSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *TwitCastingSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *TwitCastingSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Anchor ID
	// https://twitcasting.tv/username
	parts := strings.Split(targetUrl, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid url")
	}
	anchorId := parts[3]

	// 2. Fetch Page to check status
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
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

	// Check live status
	reStatus := regexp.MustCompile(`data-is-onlive="(.*?)"`)
	matchesStatus := reStatus.FindStringSubmatch(htmlStr)
	if len(matchesStatus) < 2 || matchesStatus[1] != "true" {
		return nil, fmt.Errorf("stream not found or live ended")
	}

	// 3. Get Stream Data
	streamUrl := fmt.Sprintf("https://twitcasting.tv/streamserver.php?target=%s&mode=client&player=pc_web", anchorId)
	reqStream, err := http.NewRequest("GET", streamUrl, nil)
	if err != nil {
		return nil, err
	}
	reqStream.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	reqStream.Header.Set("Referer", targetUrl)
	if s.Cookies != "" {
		reqStream.Header.Set("Cookie", s.Cookies)
	}

	respStream, err := client.Do(reqStream)
	if err != nil {
		return nil, err
	}
	defer respStream.Body.Close()

	streamBody, err := io.ReadAll(respStream.Body)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(streamBody, &jsonResult); err != nil {
		return nil, err
	}

	// 4. Extract HLS URL
	if tcHls, ok := jsonResult["tc-hls"].(map[string]interface{}); ok {
		if streams, ok := tcHls["streams"].(map[string]interface{}); ok {
			// Sort by quality: high > medium > low
			type Stream struct {
				Quality string
				Url     string
				Rank    int
			}
			var streamList []Stream
			ranks := map[string]int{"high": 0, "medium": 1, "low": 2}

			for q, u := range streams {
				if urlStr, ok := u.(string); ok {
					rank, exists := ranks[q]
					if !exists {
						rank = 99
					}
					streamList = append(streamList, Stream{Quality: q, Url: urlStr, Rank: rank})
				}
			}

			sort.Slice(streamList, func(i, j int) bool {
				return streamList[i].Rank < streamList[j].Rank
			})

			if len(streamList) > 0 {
				return &StreamInfo{Url: streamList[0].Url}, nil
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
