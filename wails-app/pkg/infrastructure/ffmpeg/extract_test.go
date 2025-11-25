package ffmpeg

import "testing"

// TestExtractSegmentPath 验证从stderr行提取分段路径。
func TestExtractSegmentPath(t *testing.T) {
    line := "Opening '/tmp/out/segment0001.ts' for writing"
    p := ExtractSegmentPath(line)
    if p != "/tmp/out/segment0001.ts" { t.Fatalf("got=%s", p) }
}

