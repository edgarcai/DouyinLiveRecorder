package app

import (
    "testing"
    "douyinrecorder/pkg/infrastructure/spider"
    "douyinrecorder/pkg/infrastructure/stream"
    "douyinrecorder/pkg/infrastructure/js"
    f "douyinrecorder/pkg/infrastructure/ffmpeg"
)

// TestPipelineLatency 验证解析→选择→参数构造延迟低于50ms（占位实现）。
func TestPipelineLatency(t *testing.T) {
    p := NewPipeline(spider.NewDouyinSpider(&js.NoopEngine{Result:""}), &stream.SimpleSelector{}, &f.Runner{})
    args, dur, err := p.PlanRecord("https://v.douyin.com/xxx", "/tmp/out/file", "mp4", 60)
    if err != nil { t.Fatal(err) }
    if len(args) == 0 { t.Fatal("empty args") }
    if dur.Milliseconds() >= 50 { t.Fatalf("latency=%dms", dur.Milliseconds()) }
}

