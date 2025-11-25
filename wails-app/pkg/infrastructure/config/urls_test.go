package config

import (
    "os"
    "path/filepath"
    "testing"
)

// TestParseURLList 验证 URL_config.ini 解析。
func TestParseURLList(t *testing.T) {
    dir := t.TempDir()
    cfg := filepath.Join(dir, "config")
    if err := os.MkdirAll(cfg, 0o755); err != nil { t.Fatal(err) }
    path := filepath.Join(cfg, "URL_config.ini")
    data := "# comment\nhttps://a.com/live # note\n\nhttps://b.com/room\n"
    if err := os.WriteFile(path, []byte(data), 0o644); err != nil { t.Fatal(err) }
    urls, err := ParseURLList(dir)
    if err != nil { t.Fatal(err) }
    if len(urls) != 2 { t.Fatalf("len=%d", len(urls)) }
    if urls[0] != "https://a.com/live" || urls[1] != "https://b.com/room" { t.Fatalf("%v", urls) }
}

