package spider

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "douyinrecorder/pkg/infrastructure/js"
    hc "douyinrecorder/pkg/infrastructure/http"
)

// TestDouyinUAHeader 验证 UA 通过 headers 传递到后端服务。
func TestDouyinUAHeader(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("User-Agent") != "UA-TEST" { t.Fatalf("ua mismatch: %s", r.Header.Get("User-Agent")) }
        w.WriteHeader(200)
        w.Write([]byte(`{"is_live":true}`))
    }))
    defer ts.Close()
    d := NewDouyinSpider(&js.NoopEngine{Result:"sig"})
    d.SetRequester(hc.NewClient())
    d.SetUA("UA-TEST")
    _, err := d.FetchRoomInfo(ts.URL)
    if err != nil { t.Fatal(err) }
}

