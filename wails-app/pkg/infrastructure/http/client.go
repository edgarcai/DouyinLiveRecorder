package httpclient

import (
    "io"
    "net/http"
    "net/url"
    "time"
)

// Requester 定义最小HTTP请求接口，供解析器依赖注入使用。
type Requester interface {
    // Get 以指定头部发起GET请求并返回状态码与响应体。
    Get(url string, headers map[string]string) (int, []byte, error)
}

// Client 提供带超时/自定义UA/Headers的HTTP客户端实现。
type Client struct{
    inner *http.Client
    defaultHeaders map[string]string
    uas []string
    uaIndex int
    maxFailures int
    failures int
}

// NewClient 构造HTTP客户端，设置合理超时与连接池。
func NewClient() *Client {
    return &Client{
        inner: &http.Client{Timeout: 15 * time.Second},
        defaultHeaders: map[string]string{
            "User-Agent": "",
            "Accept": "*/*",
        },
        uas: []string{
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36",
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Version/17.0 Safari/605.1.15",
            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/119 Safari/537.36",
        },
        maxFailures: 5,
    }
}

// Get 发起GET请求，并合并默认与外部头部。
func (c *Client) Get(url string, headers map[string]string) (int, []byte, error) {
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil { return 0, nil, err }
    // 轮换UA
    if len(c.uas) > 0 {
        ua := c.uas[c.uaIndex%len(c.uas)]
        c.uaIndex++
        req.Header.Set("User-Agent", ua)
    }
    for k, v := range c.defaultHeaders {
        if k == "User-Agent" && v == "" { continue }
        req.Header.Set(k, v)
    }
    for k, v := range headers { req.Header.Set(k, v) }
    resp, err := c.inner.Do(req)
    if err != nil { return 0, nil, err }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    if err != nil { return resp.StatusCode, nil, err }
    return resp.StatusCode, b, nil
}

// GetRetry 带重试的GET请求，指数退避（1s→3s→9s）。
func (c *Client) GetRetry(url string, headers map[string]string, attempts int) (int, []byte, error) {
    var code int
    var body []byte
    var err error
    backoff := time.Second
    for i := 0; i < attempts; i++ {
        if c.failures >= c.maxFailures { return code, body, err }
        code, body, err = c.Get(url, headers)
        if err == nil && code == 200 { c.failures = 0; return code, body, nil }
        c.failures++
        time.Sleep(backoff)
        backoff = backoff * 3
    }
    return code, body, err
}

// SetMaxFailures 设置熔断最大失败次数。
func (c *Client) SetMaxFailures(n int) { if n > 0 { c.maxFailures = n } }

// SetUAList 设置UA轮换列表。
func (c *Client) SetUAList(list []string) { if len(list) > 0 { c.uas = list; c.uaIndex = 0 } }

// SetProxy 设置HTTP代理。
func (c *Client) SetProxy(proxy string) error {
    if proxy == "" { c.inner.Transport = nil; return nil }
    u, err := url.Parse(proxy)
    if err != nil { return err }
    c.inner.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
    return nil
}
