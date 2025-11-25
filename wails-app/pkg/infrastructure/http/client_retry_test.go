package httpclient

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

// TestGetRetry 验证指数退避重试直到成功。
func TestGetRetry(t *testing.T) {
    tries := 0
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tries++
        if tries < 3 { w.WriteHeader(500); w.Write([]byte("err")); return }
        w.WriteHeader(200); w.Write([]byte("ok"))
    }))
    defer ts.Close()
    c := NewClient()
    code, body, err := c.GetRetry(ts.URL, nil, 5)
    if err != nil { t.Fatal(err) }
    if code != 200 || string(body) != "ok" || tries != 3 { t.Fatalf("unexpected code=%d body=%s tries=%d", code, string(body), tries) }
}

