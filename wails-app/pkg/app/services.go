package app

import "douyinrecorder/pkg/domain"

// SpiderProvider 定义站点解析器接口，对应现有 src/spider.py 的平台函数集合。
type SpiderProvider interface {
    // FetchRoomInfo 根据输入URL抓取并归一化直播间信息。
    FetchRoomInfo(url string) (domain.RoomInfo, error)
}

// StreamSelector 根据策略从多个播放源中选择最佳源。
type StreamSelector interface {
    // Select 依据质量偏好与可用性检测返回最佳播放源。
    Select(room domain.RoomInfo) (domain.PlaySource, error)
}

// RecorderCoordinator 负责任务编排与FFmpeg调用。
type RecorderCoordinator interface {
    // StartRecord 按指定任务启动录制流程。
    StartRecord(task domain.RecordTask) error
    // StopRecord 停止指定任务。
    StopRecord(taskID string) error
}

// PushNotifier 统一消息推送接口，兼容钉钉/微信/TG/邮箱等。
type PushNotifier interface {
    // Notify 在关键事件（开播/下播/错误）触发消息推送。
    Notify(event string, payload map[string]any) error
}

