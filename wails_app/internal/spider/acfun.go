package spider

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

type AcfunSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewAcfunSpider() *AcfunSpider {
	return &AcfunSpider{}
}

func (s *AcfunSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *AcfunSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

func (s *AcfunSpider) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *AcfunSpider) getSignParams(client *http.Client) (string, string, string, error) {
	did := fmt.Sprintf("web_%s", s.generateRandomString(16))

	params := url.Values{}
	params.Set("sid", "acfun.api.visitor")

	req, err := http.NewRequest("POST", "https://id.app.acfun.cn/rest/app/visitor/login", strings.NewReader(params.Encode()))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://live.acfun.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Cookie", fmt.Sprintf("_did=%s;", did))
	if s.Cookies != "" {
		req.Header.Set("Cookie", s.Cookies)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
		return "", "", "", err
	}

	var userId string
	if uid, ok := jsonResult["userId"].(float64); ok {
		userId = fmt.Sprintf("%.0f", uid)
	} else if uid, ok := jsonResult["userId"].(string); ok {
		userId = uid
	}

	var visitorSt string
	if st, ok := jsonResult["acfun.api.visitor_st"].(string); ok {
		visitorSt = st
	}

	return userId, did, visitorSt, nil
}

func (s *AcfunSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Author ID
	// https://live.acfun.cn/live/123456
	parts := strings.Split(strings.Split(targetUrl, "?")[0], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid url")
	}
	authorId := parts[len(parts)-1]

	client := &http.Client{}

	// 2. Check User Info
	userInfoApi := fmt.Sprintf("https://live.acfun.cn/rest/pc-direct/user/userInfo?userId=%s", authorId)
	req, err := http.NewRequest("GET", userInfoApi, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Referer", targetUrl)
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

	var jsonResult map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
		return nil, err
	}

	if profile, ok := jsonResult["profile"].(map[string]interface{}); ok {
		if _, ok := profile["liveId"]; !ok {
			return nil, fmt.Errorf("stream not found or live ended")
		}
	} else {
		return nil, fmt.Errorf("failed to get user profile")
	}

	// 3. Get Sign Params
	userId, did, visitorSt, err := s.getSignParams(client)
	if err != nil {
		return nil, fmt.Errorf("failed to get sign params: %v", err)
	}

	// 4. Get Play URL
	params := url.Values{}
	params.Set("subBiz", "mainApp")
	params.Set("kpn", "ACFUN_APP")
	params.Set("kpf", "PC_WEB")
	params.Set("userId", userId)
	params.Set("did", did)
	params.Set("acfun.api.visitor_st", visitorSt)

	data := url.Values{}
	data.Set("authorId", authorId)
	data.Set("pullStreamType", "FLV")

	playApi := fmt.Sprintf("https://api.kuaishouzt.com/rest/zt/live/web/startPlay?%s", params.Encode())
	reqPlay, err := http.NewRequest("POST", playApi, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	reqPlay.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPlay.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	reqPlay.Header.Set("Referer", targetUrl)
	reqPlay.Header.Set("Cookie", fmt.Sprintf("_did=%s;", did))
	if s.Cookies != "" {
		reqPlay.Header.Set("Cookie", s.Cookies)
	}

	respPlay, err := client.Do(reqPlay)
	if err != nil {
		return nil, err
	}
	defer respPlay.Body.Close()

	playBody, err := io.ReadAll(respPlay.Body)
	if err != nil {
		return nil, err
	}

	var jsonPlay map[string]interface{}
	if err := json.Unmarshal(playBody, &jsonPlay); err != nil {
		return nil, err
	}

	if dataMap, ok := jsonPlay["data"].(map[string]interface{}); ok {
		if videoPlayRes, ok := dataMap["videoPlayRes"].(string); ok {
			var videoRes map[string]interface{}
			if err := json.Unmarshal([]byte(videoPlayRes), &videoRes); err != nil {
				return nil, err
			}

			if liveAdaptiveManifest, ok := videoRes["liveAdaptiveManifest"].([]interface{}); ok && len(liveAdaptiveManifest) > 0 {
				if manifest, ok := liveAdaptiveManifest[0].(map[string]interface{}); ok {
					if adaptationSet, ok := manifest["adaptationSet"].(map[string]interface{}); ok {
						if representation, ok := adaptationSet["representation"].([]interface{}); ok && len(representation) > 0 {
							if rep, ok := representation[0].(map[string]interface{}); ok {
								if urlStr, ok := rep["url"].(string); ok {
									return &StreamInfo{Url: urlStr}, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found")
}
