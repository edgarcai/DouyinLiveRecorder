package app

import (
    "time"
    "douyinrecorder/pkg/domain"
)

// Pipeline 提供解析→选择→FFmpeg参数构造的一体化服务。
type Pipeline struct{
    sp  SpiderProvider
    sel StreamSelector
    run RunnerAPI
}

// RunnerAPI 是 FFmpeg 运行器接口的最小子集，用于参数构造。
type RunnerAPI interface {
    BuildArgs(task domain.RecordTask) []string
}

// NewPipeline 构造服务。
func NewPipeline(sp SpiderProvider, sel StreamSelector, run RunnerAPI) *Pipeline {
    return &Pipeline{sp: sp, sel: sel, run: run}
}

// PlanRecord 解析URL并选择源，返回FFmpeg参数；同时返回耗时用于性能监测。
func (p *Pipeline) PlanRecord(url string, outPath string, format string, segmentSec int) ([]string, time.Duration, error) {
    start := time.Now()
    room, err := p.sp.FetchRoomInfo(url)
    if err != nil { return nil, 0, err }
    src, err := p.sel.Select(room)
    if err != nil { return nil, 0, err }
    task := domain.RecordTask{Room: room, OutputPath: outPath, Format: format, SegmentSec: segmentSec}
    // 覆盖选出的URL到任务房间字段，确保 BuildArgs 使用最佳源。
    if src.URL != "" {
        task.Room.M3U8URL = src.URL
    }
    args := p.run.BuildArgs(task)
    return args, time.Since(start), nil
}

