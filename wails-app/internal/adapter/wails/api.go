package wailsadapter

import (
    "context"
    "douyinrecorder/pkg/app"
    "douyinrecorder/pkg/domain"
)

// RecorderService 提供给前端的录制相关RPC接口，占位实现。
type RecorderService struct{
    coord app.RecorderCoordinator
    ctx   context.Context
}

// NewRecorderService 构造函数。
func NewRecorderService(c app.RecorderCoordinator) *RecorderService {
    return &RecorderService{coord: c}
}

// WailsInit 在应用启动时注入上下文，用于事件推送。
func (r *RecorderService) WailsInit(ctx context.Context) { r.ctx = ctx }

// StartRecordRPC 供前端调用的录制启动方法（IPC绑定）。
func (r *RecorderService) StartRecordRPC(task domain.RecordTask) error {
    return r.coord.StartRecord(task)
}

// StopRecordRPC 供前端调用的录制停止方法。
func (r *RecorderService) StopRecordRPC(taskID string) error {
    return r.coord.StopRecord(taskID)
}
