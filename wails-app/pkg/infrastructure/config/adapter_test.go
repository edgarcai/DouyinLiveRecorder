package config

import (
    "os"
    "path/filepath"
    "testing"
)

// TestLoadAndRender 验证INI读取与${key_name}渲染向下兼容。
func TestLoadAndRender(t *testing.T) {
    dir := t.TempDir()
    cfgDir := filepath.Join(dir, "config")
    if err := os.MkdirAll(cfgDir, 0o755); err != nil { t.Fatal(err) }
    iniPath := filepath.Join(cfgDir, "config.ini")
    content := "[录制设置]\n输出目录=${key_name}/downloads\n"
    if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil { t.Fatal(err) }

    adapter, err := LoadConfig(dir)
    if err != nil { t.Fatalf("load error: %v", err) }

    raw, ok := adapter.GetRaw("录制设置.输出目录")
    if !ok { t.Fatalf("missing key") }
    if raw != "${key_name}/downloads" { t.Fatalf("raw mismatch: %s", raw) }

    // 渲染存在的占位符
    got, err := adapter.RenderValue("录制设置.输出目录", map[string]string{"key_name": "ROOT"})
    if err != nil { t.Fatalf("render error: %v", err) }
    if got != "ROOT/downloads" { t.Fatalf("render mismatch: %s", got) }

    // 渲染缺失占位符：保持原值向下兼容
    got2, err := adapter.RenderValue("录制设置.输出目录", map[string]string{})
    if err != nil { t.Fatalf("render error: %v", err) }
    if got2 != "${key_name}/downloads" { t.Fatalf("fallback mismatch: %s", got2) }
}

