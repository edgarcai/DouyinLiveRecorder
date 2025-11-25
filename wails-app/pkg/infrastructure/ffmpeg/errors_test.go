package ffmpeg

import "testing"

// TestClassify 验证错误分类规则。
func TestClassify(t *testing.T) {
    if ClassifyFFmpegError("Connection reset by peer") != "network" { t.Fatal("network") }
    if ClassifyFFmpegError("protocol not found") != "protocol" { t.Fatal("protocol") }
    if ClassifyFFmpegError("No space left on device") != "write" { t.Fatal("write") }
}

