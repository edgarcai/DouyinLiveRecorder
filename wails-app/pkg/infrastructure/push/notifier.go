package push

// Notifier 提供统一的消息推送占位实现。
// 后续将对接钉钉/微信/TG/邮箱等具体渠道，并支持配置控制。
type Notifier struct{}

// Notify 触发消息推送占位，实现接口期望的行为。
func (n *Notifier) Notify(event string, payload map[string]any) error {
    // 占位：后续接入各渠道SDK或HTTP API。
    return nil
}

