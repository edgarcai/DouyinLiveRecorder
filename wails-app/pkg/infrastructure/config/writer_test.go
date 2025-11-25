package config

import (
    "os"
    "path/filepath"
    "testing"
)

// TestUpdateExistingKey 验证更新现有键值生效。
func TestUpdateExistingKey(t *testing.T) {
    dir := t.TempDir()
    cfg := filepath.Join(dir, "config")
    if err := os.MkdirAll(cfg, 0o755); err != nil { t.Fatal(err) }
    path := filepath.Join(cfg, "config.ini")
    data := "[录制设置]\n输出目录=/tmp/downloads\n"
    if err := os.WriteFile(path, []byte(data), 0o644); err != nil { t.Fatal(err) }
    if err := Update(dir, "录制设置.输出目录", "/data/downloads"); err != nil { t.Fatal(err) }
    a, err := LoadConfig(dir)
    if err != nil { t.Fatal(err) }
    v, ok := a.GetRaw("录制设置.输出目录")
    if !ok || v != "/data/downloads" { t.Fatalf("got=%s ok=%v", v, ok) }
}

// TestAppendNewKey 验证在节内不存在的键追加写入。
func TestAppendNewKey(t *testing.T) {
    dir := t.TempDir()
    cfg := filepath.Join(dir, "config")
    if err := os.MkdirAll(cfg, 0o755); err != nil { t.Fatal(err) }
    path := filepath.Join(cfg, "config.ini")
    data := "[录制设置]\n输出目录=/tmp/downloads\n"
    if err := os.WriteFile(path, []byte(data), 0o644); err != nil { t.Fatal(err) }
    if err := Update(dir, "录制设置.视频保存格式", "mp4"); err != nil { t.Fatal(err) }
    a, err := LoadConfig(dir)
    if err != nil { t.Fatal(err) }
    v, ok := a.GetRaw("录制设置.视频保存格式")
    if !ok || v != "mp4" { t.Fatalf("got=%s ok=%v", v, ok) }
}
