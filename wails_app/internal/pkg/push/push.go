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

func (p *PushService) UpdateConfig(cfg *config.PushSettings) {
	p.config = cfg
}

func (p *PushService) Send(title, content string) {
	// Split channels by comma (support both English and Chinese comma)
	channelsStr := strings.ReplaceAll(p.config.PushChannels, "，", ",")
	channels := strings.Split(channelsStr, ",")
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

	urlsStr := strings.ReplaceAll(p.config.DingTalkUrl, "，", ",")
	urls := strings.Split(urlsStr, ",")

	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

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

		p.postJson(url, payload)
	}
}

func (p *PushService) WeChat(title, content string) {
	// Xizhi
	if p.config.WeChatUrl == "" {
		return
	}

	urlsStr := strings.ReplaceAll(p.config.WeChatUrl, "，", ",")
	urls := strings.Split(urlsStr, ",")

	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		payload := map[string]string{
			"title":   title,
			"content": content,
		}
		p.postJson(url, payload)
	}
}

func (p *PushService) Bark(title, content string) {
	if p.config.BarkUrl == "" {
		return
	}

	urlsStr := strings.ReplaceAll(p.config.BarkUrl, "，", ",")
	urls := strings.Split(urlsStr, ",")

	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		// Bark URL format: https://api.day.app/key/
		// We can post JSON
		payload := map[string]interface{}{
			"title":     title,
			"body":      content,
			"level":     p.config.BarkLevel,
			"sound":     p.config.BarkSound,
			"badge":     1,
			"autoCopy":  1,
			"isArchive": 1,
			// "icon": "", // Not in config yet
			// "group": "", // Not in config yet
			// "url": "", // Not in config yet
		}
		p.postJson(url, payload)
	}
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

	receiversStr := strings.ReplaceAll(p.config.ReceiverEmail, "，", ",")
	receivers := strings.Split(receiversStr, ",")
	var validReceivers []string
	for _, r := range receivers {
		r = strings.TrimSpace(r)
		if r != "" {
			validReceivers = append(validReceivers, r)
		}
	}

	if len(validReceivers) == 0 {
		return
	}

	// Header
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("=?UTF-8?B?%s?= <%s>", base64.StdEncoding.EncodeToString([]byte(p.config.SenderName)), p.config.SenderEmail)
	// To header usually shows the first receiver or all, but for privacy or simplicity, let's just show the first one or leave it generic if multiple.
	// Actually, standard practice is to list them or send individual emails.
	// Python sends one email with multiple recipients in 'To' header if I recall correctly?
	// Python: message['To'] = receivers[0] if len(receivers) == 1 else ... wait, Python code:
	// if len(receivers) == 1: message['To'] = receivers[0]
	// It doesn't set 'To' header for multiple? That might trigger spam filters.
	// Let's set To to the first one or a string of all.
	header["To"] = strings.Join(validReceivers, ",")
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
		err = sendMailSSL(addr, auth, p.config.SenderEmail, validReceivers, []byte(message))
	} else {
		err = smtp.SendMail(addr, auth, p.config.SenderEmail, validReceivers, []byte(message))
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

	urlsStr := strings.ReplaceAll(p.config.NtfyUrl, "，", ",")
	urls := strings.Split(urlsStr, ",")

	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		// Split server and topic
		// Python: server, topic = _api.rsplit('/', maxsplit=1)
		lastSlashIndex := strings.LastIndex(url, "/")
		if lastSlashIndex == -1 {
			continue
		}
		server := url[:lastSlashIndex]
		topic := url[lastSlashIndex+1:]

		payload := map[string]interface{}{
			"topic":    topic,
			"title":    title,
			"message":  content,
			"tags":     []string{p.config.NtfyTag},
			"priority": 3,
			"markdown": false,
		}

		p.postJson(server, payload)
	}
}

func (p *PushService) PushPlus(title, content string) {
	if p.config.PushPlusToken == "" {
		return
	}

	tokensStr := strings.ReplaceAll(p.config.PushPlusToken, "，", ",")
	tokens := strings.Split(tokensStr, ",")

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		url := "https://www.pushplus.plus/send"
		payload := map[string]string{
			"token":   token,
			"title":   title,
			"content": content,
		}
		p.postJson(url, payload)
	}
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
