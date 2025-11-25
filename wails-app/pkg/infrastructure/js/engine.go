package js

// Engine 定义 JS 执行器接口，便于后续接入 goja 或外部 Node。
type Engine interface {
    // Eval 执行脚本并返回字符串结果。
    Eval(script string) (string, error)
    // EvalWithArgs 执行脚本并传入参数，返回结果（用于签名计算）。
    EvalWithArgs(script string, args []string) (string, error)
}

// NoopEngine 提供占位实现，返回固定值。
type NoopEngine struct{ Result string }

// Eval 返回预设结果或空串。
func (n *NoopEngine) Eval(script string) (string, error) { return n.Result, nil }

// EvalWithArgs 返回预设结果或空串。
func (n *NoopEngine) EvalWithArgs(script string, args []string) (string, error) { return n.Result, nil }
