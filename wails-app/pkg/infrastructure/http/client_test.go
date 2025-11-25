package httpclient

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

// TestClientGet 验证GET请求与头部合并行为。
func TestClientGet(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-Test") != "ok" { t.Fatalf("missing header") }
        w.WriteHeader(200)
        w.Write([]byte("{}"))
    }))
    defer ts.Close()
    c := NewClient()
    code, body, err := c.Get(ts.URL, map[string]string{"X-Test": "ok"})
    if err != nil { t.Fatal(err) }
    if code != 200 || string(body) != "{}" { t.Fatalf("unexpected %d %s", code, string(body)) }
}

