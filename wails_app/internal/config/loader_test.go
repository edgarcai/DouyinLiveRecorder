package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
[录制设置]
language(zh_cn/en) = zh_cn
是否跳过代理检测(是/否) = 是
`)
	tmpfile, err := os.CreateTemp("", "config.ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading
	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.RecordingSettings.Language != "zh_cn" {
		t.Errorf("Expected Language zh_cn, got %s", cfg.RecordingSettings.Language)
	}
	if cfg.RecordingSettings.SkipProxyCheck != "是" {
		t.Errorf("Expected SkipProxyCheck 是, got %s", cfg.RecordingSettings.SkipProxyCheck)
	}
}
