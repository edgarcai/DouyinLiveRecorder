package wailsadapter

import (
    "os"
    "path/filepath"
    "testing"
)

// TestAuthServiceReadWriteMasked 验证读写与遮蔽输出。
func TestAuthServiceReadWriteMasked(t *testing.T) {
    dir := t.TempDir()
    cfg := filepath.Join(dir, "config")
    if err := os.MkdirAll(cfg, 0o755); err != nil { t.Fatal(err) }
    path := filepath.Join(cfg, "config.ini")
    data := "[登录]\n抖音Token=ABCDEF123456\n"
    if err := os.WriteFile(path, []byte(data), 0o644); err != nil { t.Fatal(err) }

    svc, err := NewAuthService(dir)
    if err != nil { t.Fatal(err) }
    v, ok := svc.GetKeyRPC("登录.抖音Token")
    if !ok || v != "ABCDEF123456" { t.Fatalf("unexpected: %v %v", v, ok) }
    m, ok := svc.GetKeyMaskedRPC("登录.抖音Token")
    if !ok || m != "****3456" { t.Fatalf("masked: %v", m) }
    if err := svc.SetKeyRPC("登录.抖音Token", "NEWVALUE9999"); err != nil { t.Fatal(err) }
    v2, ok := svc.GetKeyRPC("登录.抖音Token")
    if !ok || v2 != "NEWVALUE9999" { t.Fatalf("updated: %v", v2) }
}

