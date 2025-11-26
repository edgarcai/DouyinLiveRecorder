package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type JDSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewJDSpider() *JDSpider {
	return &JDSpider{}
}

func (s *JDSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *JDSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *JDSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Fetch to get redirect URL and Author ID
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	req.Header.Set("origin", "https://lives.jd.com")
	req.Header.Set("referer", "https://lives.jd.com/")
	req.Header.Set("x-referer-page", "https://lives.jd.com/")
	if s.Cookies != "" {
		req.Header.Set("Cookie", s.Cookies)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	finalUrl := resp.Request.URL.String()

	// 2. Extract Author ID
	u, err := url.Parse(finalUrl)
	if err != nil {
		return nil, err
	}
	authorId := u.Query().Get("authorId")

	if authorId == "" {
		// Try regex if not in query params
		re := regexp.MustCompile(`#/(.*?)\?origin`)
		matches := re.FindStringSubmatch(finalUrl)
		if len(matches) < 2 {
			return nil, fmt.Errorf("failed to extract author id")
		}
		// If found via regex, it might be liveId, but let's assume we need authorId for API
		// Python code logic: if not authorId, try regex for liveId, then return result with anchor_name=jd_{liveId} and is_live=False?
		// Wait, Python code says if not authorId, it tries to get liveId and returns result immediately (implying not live or just basic info).
		// But to get stream url we need to call API.
		// Let's assume we need authorId.
		return nil, fmt.Errorf("author id not found")
	}

	// 3. Call API
	// https://api.m.jd.com/talent_head_findTalentMsg
	apiUrl := "https://api.m.jd.com/talent_head_findTalentMsg"

	bodyData := fmt.Sprintf(`{"authorId":"%s","monitorSource":"1","userId":""}`, authorId)

	params := url.Values{}
	params.Set("functionId", "talent_head_findTalentMsg")
	params.Set("appid", "dr_detail")
	params.Set("body", bodyData)

	reqApi, err := http.NewRequest("POST", apiUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	reqApi.Header.Set("User-Agent", "ios/7.830 (ios 17.0; ; iPhone 15 (A2846/A3089/A3090/A3092))")
	reqApi.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqApi.Header.Set("origin", "https://lives.jd.com")
	reqApi.Header.Set("referer", "https://lives.jd.com/")
	if s.Cookies != "" {
		reqApi.Header.Set("Cookie", s.Cookies)
	}

	respApi, err := client.Do(reqApi)
	if err != nil {
		return nil, err
	}
	defer respApi.Body.Close()

	bodyBytes, err := io.ReadAll(respApi.Body)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
		return nil, err
	}

	// 4. Extract Stream URL
	if result, ok := jsonResult["result"].(map[string]interface{}); ok {
		if livingRoomJump, ok := result["livingRoomJump"].(map[string]interface{}); ok {
			if params, ok := livingRoomJump["params"].(map[string]interface{}); ok {
				if playUrl, ok := params["playUrl"].(string); ok && playUrl != "" {
					return &StreamInfo{Url: playUrl}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
