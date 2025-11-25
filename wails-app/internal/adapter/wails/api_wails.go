//go:build wails

package wailsadapter

import (
    "context"
    "douyinrecorder/pkg/domain"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsInit 注入上下文，用于事件推送（仅在 wails 构建包含）。
func (r *RecorderService) WailsInit(ctx context.Context) { r.ctx = ctx }

// StartRecordRPC 扩展事件推送。
func (r *RecorderService) StartRecordRPC(task domain.RecordTask) error {
    if err := r.coord.StartRecord(task); err != nil { return err }
    if r.ctx != nil { runtime.EventsEmit(r.ctx, "record_started", task) }
    return nil
}

// StopRecordRPC 扩展事件推送。
func (r *RecorderService) StopRecordRPC(taskID string) error {
    if err := r.coord.StopRecord(taskID); err != nil { return err }
    if r.ctx != nil { runtime.EventsEmit(r.ctx, "record_stopped", taskID) }
    return nil
}

