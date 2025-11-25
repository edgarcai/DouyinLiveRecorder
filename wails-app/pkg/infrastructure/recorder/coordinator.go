package recorder

import (
    "context"
    "sync"
    "douyinrecorder/pkg/domain"
    "douyinrecorder/pkg/infrastructure/ffmpeg"
)

// Coordinator 提供最小可用的录制编排占位实现。
type Coordinator struct {
    mu    sync.Mutex
    tasks map[string]domain.RecordTask
    run   ffmpeg.RunnerAPI
}

// NewCoordinator 构造并初始化任务存储。
func NewCoordinator() *Coordinator {
    return &Coordinator{tasks: make(map[string]domain.RecordTask), run: &ffmpeg.Runner{}}
}

// StartRecord 启动录制占位：登记任务并构造 FFmpeg 参数（暂不执行）。
func (c *Coordinator) StartRecord(task domain.RecordTask) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    key := task.Room.Platform + ":" + task.Room.AnchorName
    c.tasks[key] = task
    ctx := context.Background()
    _, _ = c.run.Start(ctx, task)
    return nil
}

// StopRecord 停止录制占位：移除任务登记。
func (c *Coordinator) StopRecord(taskID string) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.tasks, taskID)
    return nil
}

// RegisterEvents 注册录制过程事件回调（段事件与错误事件）。
func (c *Coordinator) RegisterEvents(onSegment func(string), onError func(string, string)) {
    // 尝试调用运行器的事件设置方法。
    type extended interface {
        SetOnSegment(func(string))
        SetOnError(func(string, string))
    }
    if x, ok := c.run.(extended); ok {
        x.SetOnSegment(onSegment)
        x.SetOnError(onError)
    }
}
