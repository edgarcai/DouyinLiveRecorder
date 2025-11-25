package crypto

import (
	"strings"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	params := "aid=6383&app_name=douyin_web&live_id=1&device_platform=web&language=zh-CN"
	ua := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.5845.97 Safari/537.36 Core/1.116.567.400 QQBrowser/19.7.6764.400"

	sig := GenerateSignature(params, ua)

	if len(sig) == 0 {
		t.Error("Signature should not be empty")
	}

	if !strings.HasSuffix(sig, "=") {
		t.Error("Signature should end with =")
	}

	// The signature length varies but usually has a specific range.
	// Based on the Python code, it produces a substantial string.
	if len(sig) < 20 {
		t.Errorf("Signature too short: %s", sig)
	}
}
