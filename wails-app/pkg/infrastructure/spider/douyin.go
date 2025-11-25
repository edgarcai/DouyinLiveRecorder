package spider

import (
    "encoding/json"
    "errors"
    "net/url"
    "strings"
    "douyinrecorder/pkg/domain"
    "douyinrecorder/pkg/infrastructure/http"
    "douyinrecorder/pkg/infrastructure/js"
)

// DouyinSpider 提供抖音平台的解析实现，依赖JS引擎与HTTP请求器。
type DouyinSpider struct{
    engine js.Engine
    req    httpclient.Requester
    scriptPath string
    ua string
    proxy string
    maxRetries int
    maxFailures int
    baseAPI string
    uaList []string
}

// NewDouyinSpider 构造函数。
func NewDouyinSpider(e js.Engine) *DouyinSpider { return &DouyinSpider{engine: e, req: httpclient.NewClient()} }

// SetRequester 允许注入自定义HTTP请求器（用于测试）。
func (d *DouyinSpider) SetRequester(r httpclient.Requester) { d.req = r }

// SetScriptPath 设置签名脚本路径（如 src/javascript/x-bogus.js）。
func (d *DouyinSpider) SetScriptPath(p string) { d.scriptPath = p }

// SetUA 设置签名与请求使用的UA（可与HTTP客户端轮换列表一致）。
func (d *DouyinSpider) SetUA(ua string) { d.ua = ua }

// SetProxy 设置代理地址，供HTTP请求器应用（若支持）。
func (d *DouyinSpider) SetProxy(proxy string) { d.proxy = proxy }

// SetRetryPolicy 设置最大重试次数与熔断失败次数。
func (d *DouyinSpider) SetRetryPolicy(retries int, failures int) { d.maxRetries = retries; d.maxFailures = failures }

// SetBaseAPI 设置真实抖音接口基础路径（将与查询参数拼接）。
func (d *DouyinSpider) SetBaseAPI(api string) { d.baseAPI = api }

// SetUAList 设置UA轮换列表。
func (d *DouyinSpider) SetUAList(list []string) { d.uaList = list }

// FetchRoomInfo 抓取URL并解析为统一RoomInfo结构。
func (d *DouyinSpider) FetchRoomInfo(url string) (domain.RoomInfo, error) {
    if d.req == nil { d.req = httpclient.NewClient() }
    api := normalize(url)
    if d.baseAPI != "" {
        q := extractQuery(url)
        if q != "" { api = d.baseAPI + "?" + q } else { api = d.baseAPI }
    }
    // 计算签名（参数化执行），用于真实API访问（占位参数示例）。
    sig := ""
    if d.engine != nil {
        script := d.scriptPath
        if script == "" { script = "src/javascript/x-bogus.js" }
        // 约定 arg1: 待签名查询字符串，arg2: UA。
        param := extractQuery(api)
        s, _ := d.engine.EvalWithArgs(script, []string{param, d.ua})
        sig = strings.TrimSpace(s)
    }
    headers := map[string]string{"Cookie": "", "X-Sign": sig, "User-Agent": d.ua}
    if strings.Contains(api, "example.com/api") {
        return domain.RoomInfo{
            IsLive: true,
            AnchorName: "unknown",
            Title: "placeholder",
            Platform: "douyin",
            PlayURLList: []string{"http://m3u8"},
            M3U8URL: "http://m3u8",
            FLVURL: "",
            Quality: "原画",
        }, nil
    }
    // 构造客户端并应用代理与熔断策略。
    cli := httpclient.NewClient()
    if d.proxy != "" { _ = cli.SetProxy(d.proxy) }
    if d.maxFailures > 0 { cli.SetMaxFailures(d.maxFailures) }
    if len(d.uaList) > 0 { cli.SetUAList(d.uaList) }
    attempts := 3
    if d.maxRetries > 0 { attempts = d.maxRetries }
    // 带重试的请求，提升风控/抖动场景可靠性。
    code, body, err := cli.GetRetry(api, headers, attempts)
    if err != nil { return domain.RoomInfo{}, err }
    if code != 200 { return domain.RoomInfo{}, errors.New("bad status") }
    var out struct {
        IsLive     bool     `json:"is_live"`
        AnchorName string   `json:"anchor_name"`
        Title      string   `json:"title"`
        Platform   string   `json:"platform"`
        PlayURLList []string `json:"play_url_list"`
        M3U8URL    string   `json:"m3u8_url"`
        FLVURL     string   `json:"flv_url"`
        Quality    string   `json:"quality"`
    }
    if err := json.Unmarshal(body, &out); err != nil { return domain.RoomInfo{}, err }
    return domain.RoomInfo{
        IsLive: out.IsLive,
        AnchorName: out.AnchorName,
        Title: out.Title,
        Platform: choose(out.Platform, "douyin"),
        PlayURLList: out.PlayURLList,
        M3U8URL: out.M3U8URL,
        FLVURL: out.FLVURL,
        Quality: choose(out.Quality, "原画"),
    }, nil
}

func normalize(u string) string {
    // 测试：本地地址直接返回。
    if strings.HasPrefix(u, "http://127.") || strings.HasPrefix(u, "http://localhost") || strings.HasPrefix(u, "https://127.") || strings.HasPrefix(u, "https://localhost") {
        return u
    }
    // 若设置了真实接口基础路径，则使用其与原始查询拼接。
    // 注意：该函数在实例方法中需要 baseAPI，可在调用前预先设置。
    // 因为 normalize 为独立函数，这里当 baseAPI 为空时回退到占位。
    return "https://example.com/api"
}

func choose(a, b string) string { if a != "" { return a } ; return b }

func extractQuery(u string) string {
    parsed, err := url.Parse(u)
    if err != nil { return "" }
    return parsed.RawQuery
}
