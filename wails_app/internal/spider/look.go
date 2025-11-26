package spider

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type LookSpider struct {
	ProxyUrl string
	Cookies  string
}

func NewLookSpider() *LookSpider {
	return &LookSpider{}
}

func (s *LookSpider) SetProxy(proxyUrl string) {
	s.ProxyUrl = proxyUrl
}

func (s *LookSpider) SetCookies(cookies string) {
	s.Cookies = cookies
}

// Encryption constants
const (
	lookModulus   = "00e0b509f6259df8642dbc35662901477df22677ec152b5ff68ace615bb7b725152b3ab17a876aea8a5aa76d2e417629ec4ee341f56135fccf695280104e0312ecbda92557c93870114af6c9d05c4f7f0c3685b7a46bee255932575cce10b424d813cfe4875d3e82047b97ddef52741d546b8e289dc6935b3ece0462db0a22b8e7"
	lookNonce     = "0CoJUm6Qyw8W8jud"
	lookPublicKey = "010001"
)

func (s *LookSpider) createSecretKey(size int) []byte {
	charset := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|;:,.<>?"
	b := make([]byte, size)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return b
}

func (s *LookSpider) aesEncrypt(text []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := []byte("0102030405060708")
	// PKCS7 Padding
	padding := aes.BlockSize - len(text)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	text = append(text, padtext...)

	ciphertext := make([]byte, len(text))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, text)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *LookSpider) rsaEncrypt(text []byte, pubKey string, mod string) string {
	// Reverse text
	reversed := make([]byte, len(text))
	for i, v := range text {
		reversed[len(text)-1-i] = v
	}

	textInt := new(big.Int).SetBytes(reversed)
	pubKeyInt, _ := new(big.Int).SetString(pubKey, 16)
	modInt, _ := new(big.Int).SetString(mod, 16)

	encryptedInt := new(big.Int).Exp(textInt, pubKeyInt, modInt)
	return fmt.Sprintf("%0256x", encryptedInt)
}

func (s *LookSpider) getSecretData(data interface{}) (string, string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	secKey := s.createSecretKey(16)

	// First AES encryption with nonce
	enc1, err := s.aesEncrypt(jsonBytes, []byte(lookNonce))
	if err != nil {
		return "", "", err
	}

	// Second AES encryption with secKey
	encText, err := s.aesEncrypt([]byte(enc1), secKey)
	if err != nil {
		return "", "", err
	}

	// RSA encryption of secKey
	encSecKey := s.rsaEncrypt(secKey, lookPublicKey, lookModulus)

	return encText, encSecKey, nil
}

func (s *LookSpider) GetStreamUrl(targetUrl string) (*StreamInfo, error) {
	// 1. Extract Room ID
	re := regexp.MustCompile(`live\?id=(.*?)&`)
	matches := re.FindStringSubmatch(targetUrl)
	if len(matches) < 2 {
		// Try without &
		re2 := regexp.MustCompile(`live\?id=(.*)`)
		matches = re2.FindStringSubmatch(targetUrl)
		if len(matches) < 2 {
			return nil, fmt.Errorf("failed to extract room id")
		}
	}
	roomId := matches[1]

	// 2. Prepare encrypted params
	params, encSecKey, err := s.getSecretData(map[string]string{"liveRoomNo": roomId})
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %v", err)
	}

	// 3. Call API
	apiUrl := "https://api.look.163.com/weapi/livestream/room/get/v3"
	data := url.Values{}
	data.Set("params", params)
	data.Set("encSecKey", encSecKey)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://look.163.com/")
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

	// 4. Extract Stream URL
	if dataMap, ok := jsonResult["data"].(map[string]interface{}); ok {
		if liveStatus, ok := dataMap["liveStatus"].(float64); ok && liveStatus == 1 {
			if roomInfo, ok := dataMap["roomInfo"].(map[string]interface{}); ok {
				if liveType, ok := roomInfo["liveType"].(float64); ok && liveType == 1 {
					return nil, fmt.Errorf("audio only live not supported yet")
				}
				if liveUrl, ok := roomInfo["liveUrl"].(string); ok {
					return &StreamInfo{Url: liveUrl}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("stream not found or live ended")
}
