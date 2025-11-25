package spider

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "douyinrecorder/pkg/infrastructure/js"
    hc "douyinrecorder/pkg/infrastructure/http"
)

// TestDouyinSpiderWithRequester 使用httptest验证JSON解析为RoomInfo。
func TestDouyinSpiderWithRequester(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write([]byte(`{"is_live":true,"anchor_name":"a","title":"t","platform":"douyin","play_url_list":["http://m3u8"],"m3u8_url":"http://m3u8","flv_url":"","quality":"原画"}`))
    }))
    defer ts.Close()
    d := NewDouyinSpider(&js.NoopEngine{Result:"sig"})
    d.SetRequester(hc.NewClient())
    info, err := d.FetchRoomInfo(ts.URL)
    if err != nil { t.Fatal(err) }
    if !info.IsLive || info.Platform != "douyin" || info.M3U8URL == "" { t.Fatalf("unexpected %+v", info) }
}

