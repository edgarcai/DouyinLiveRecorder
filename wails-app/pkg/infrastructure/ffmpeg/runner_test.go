package ffmpeg

import (
    "testing"
    "douyinrecorder/pkg/domain"
)

// TestBuildArgs 验证参数构造符合期望（含分段与格式）。
func TestBuildArgs(t *testing.T) {
    r := &Runner{}
    task := domain.RecordTask{
        Room: domain.RoomInfo{M3U8URL: "http://m3u8", PlayURLList: []string{"http://flv"}},
        OutputPath: "/tmp/out/file",
        Format: "mp4",
        SegmentSec: 120,
    }
    args := r.BuildArgs(task)
    if len(args) == 0 { t.Fatal("empty args") }
    // 简要检查包含关键片段
    expected := []string{"-i", "http://m3u8", "-f", "segment", "-segment_time", "120"}
    for _, e := range expected {
        found := false
        for _, a := range args { if a == e { found = true; break } }
        if !found { t.Fatalf("missing %s in %v", e, args) }
    }
}
