package wailsadapter

import (
    "douyinrecorder/pkg/app"
)

// PipelineService 提供前端调用的一体化规划接口。
type PipelineService struct{ p *app.Pipeline }

// NewPipelineService 构造服务。
func NewPipelineService(p *app.Pipeline) *PipelineService { return &PipelineService{p: p} }

// PlanRecordRPC 返回参数与耗时（纳秒）。
func (s *PipelineService) PlanRecordRPC(url, outPath, format string, segmentSec int) ([]string, int64, error) {
    args, dur, err := s.p.PlanRecord(url, outPath, format, segmentSec)
    if err != nil { return nil, 0, err }
    return args, int64(dur), nil
}

