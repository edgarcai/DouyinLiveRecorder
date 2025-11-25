package push

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"
	"wails_app/internal/config"
)

type PushService struct {
	config *config.PushSettings
}

func NewPushService(cfg *config.PushSettings) *PushService {
	return &PushService{config: cfg}
}

func (p *PushService) Send(title, content string) {
	// Split channels by comma
	channels := strings.Split(p.config.PushChannels, ",")
	for _, channel := range channels {
		channel = strings.TrimSpace(channel)
		switch channel {
		case "dingtalk":
			p.DingTalk(title, content)
		case "wechat":
			p.WeChat(title, content)
		case "bark":
			p.Bark(title, content)
		case "tg":
			p.Telegram(title, content)
		case "email":
			p.Email(title, content)
		case "ntfy":
			p.Ntfy(title, content)
		case "pushplus":
			p.PushPlus(title, content)
		}
	}
}

func (p *PushService) DingTalk(title, content string) {
	if p.config.DingTalkUrl == "" {
		return
	}

	// DingTalk expects markdown or text. Python used text.
	// Python logic:
	/*
	   json_data = {
	       'msgtype': 'text',
	       'text': {'content': content},
	       "at": {"atMobiles": [number], "isAtAll": is_atall},
	   }
	*/

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n%s", title, content),
		},
		"at": map[string]interface{}{
			"atMobiles": []string{p.config.DingTalkPhone},
			"isAtAll":   p.config.DingTalkAtAll == "是",
		},
	}

	p.postJson(p.config.DingTalkUrl, payload)
}

func (p *PushService) WeChat(title, content string) {
	// Xizhi
	if p.config.WeChatUrl == "" {
		return
	}
	payload := map[string]string{
		"title":   title,
		"content": content,
	}
	p.postJson(p.config.WeChatUrl, payload)
}

func (p *PushService) Bark(title, content string) {
	if p.config.BarkUrl == "" {
		return
	}
	// Bark URL format: https://api.day.app/key/
	// We can post JSON
	payload := map[string]interface{}{
		"title":     title,
		"body":      content,
		"level":     p.config.BarkLevel,
		"sound":     p.config.BarkSound,
		"autoCopy":  1,
		"isArchive": 1,
	}
	p.postJson(p.config.BarkUrl, payload)
}

func (p *PushService) Telegram(title, content string) {
	if p.config.TgToken == "" || p.config.TgChatId == "" {
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", p.config.TgToken)
	payload := map[string]interface{}{
		"chat_id": p.config.TgChatId,
		"text":    fmt.Sprintf("%s\n%s", title, content),
	}
	p.postJson(url, payload)
}

func (p *PushService) Email(title, content string) {
	// Basic SMTP implementation
	if p.config.SmtpServer == "" {
		return
	}

	auth := smtp.PlainAuth("", p.config.EmailAccount, p.config.EmailPassword, p.config.SmtpServer)

	// Header
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("=?UTF-8?B?%s?= <%s>", base64.StdEncoding.EncodeToString([]byte(p.config.SenderName)), p.config.SenderEmail)
	header["To"] = p.config.ReceiverEmail
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(title)))
	header["Content-Type"] = "text/plain; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + content

	addr := fmt.Sprintf("%s:%s", p.config.SmtpServer, p.config.SmtpPort)

	var err error
	if p.config.SmtpSsl == "是" {
		// SMTP SSL (usually port 465) requires custom dialer
		// Simplified for now, standard smtp.SendMail uses STARTTLS if supported on 587
		// For 465 SSL, we need tls.Dial
		err = sendMailSSL(addr, auth, p.config.SenderEmail, []string{p.config.ReceiverEmail}, []byte(message))
	} else {
		err = smtp.SendMail(addr, auth, p.config.SenderEmail, []string{p.config.ReceiverEmail}, []byte(message))
	}

	if err != nil {
		fmt.Printf("Email send error: %v\n", err)
	}
}

func sendMailSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Simple SSL implementation
	host, _, _ := strings.Cut(addr, ":")
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *PushService) Ntfy(title, content string) {
	if p.config.NtfyUrl == "" {
		return
	}
	// ntfy.sh/topic
	payload := map[string]interface{}{
		"topic":   strings.TrimPrefix(p.config.NtfyUrl, "https://ntfy.sh/"), // Simplified extraction
		"title":   title,
		"message": content,
		"tags":    []string{p.config.NtfyTag},
	}
	// If url is full url, we might need to parse it better or just use it as endpoint if it's self-hosted
	// Python code splits server and topic.
	// Let's assume NtfyUrl is the full topic URL for simplicity or handle it like python

	targetUrl := p.config.NtfyUrl
	// If it's just the base url, we might need topic. But config says "NtfyUrl", usually full path.

	// Python: server, topic = _api.rsplit('/', maxsplit=1)
	// json_data = { "topic": topic ... }
	// req = Request(server, ...)
	// This implies posting to root server with topic in body.

	// Let's just POST to the URL directly which ntfy also supports (publish via POST)
	// If we post to https://ntfy.sh/topic, body is message.
	// Or we can post JSON to https://ntfy.sh

	// Let's try posting JSON to the URL provided.
	p.postJson(targetUrl, payload)
}

func (p *PushService) PushPlus(title, content string) {
	if p.config.PushPlusToken == "" {
		return
	}
	url := "https://www.pushplus.plus/send"
	payload := map[string]string{
		"token":   p.config.PushPlusToken,
		"title":   title,
		"content": content,
	}
	p.postJson(url, payload)
}

func (p *PushService) postJson(url string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("JSON marshal error: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("NewRequest error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP post error to %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	// We could check response status here
}
