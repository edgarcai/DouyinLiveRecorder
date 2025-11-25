package wailsadapter

import (
    "douyinrecorder/pkg/infrastructure/config"
)

// ConfigService 供前端调用的配置服务桥接，封装渲染与原值读取。
type ConfigService struct{
    svc *config.Service
}

// NewConfigService 构造桥接服务。
func NewConfigService(root string) (*ConfigService, error) {
    s, err := config.NewService(root)
    if err != nil { return nil, err }
    return &ConfigService{svc: s}, nil
}

// GetRawRPC 读取原始值。
func (c *ConfigService) GetRawRPC(key string) (string, bool) { return c.svc.GetRaw(key) }

// RenderRPC 渲染占位符后的值。
func (c *ConfigService) RenderRPC(key string, ctx map[string]string) (string, error) {
    return c.svc.Render(key, ctx)
}

