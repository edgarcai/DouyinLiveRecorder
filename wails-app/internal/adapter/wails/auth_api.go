package wailsadapter

import "douyinrecorder/pkg/infrastructure/config"

// AuthService 提供登录相关的配置读写桥接。
type AuthService struct{
    cfg *config.Service
    root string
}

// NewAuthService 构造服务。
func NewAuthService(root string) (*AuthService, error) {
    s, err := config.NewService(root)
    if err != nil { return nil, err }
    return &AuthService{cfg: s, root: root}, nil
}

// GetKeyRPC 读取 composite 键的原值（敏感键建议在前端进行遮蔽显示）。
func (a *AuthService) GetKeyRPC(key string) (string, bool) { return a.cfg.GetRaw(key) }

// SetKeyRPC 更新 composite 键的值。
func (a *AuthService) SetKeyRPC(key, value string) error { return a.cfg.SetRaw(a.root, key, value) }

// GetKeyMaskedRPC 返回遮蔽后的敏感值，仅保留末尾4位可见。
func (a *AuthService) GetKeyMaskedRPC(key string) (string, bool) {
    v, ok := a.cfg.GetRaw(key)
    if !ok { return "", false }
    if len(v) <= 4 { return "****", true }
    return "****" + v[len(v)-4:], true
}
