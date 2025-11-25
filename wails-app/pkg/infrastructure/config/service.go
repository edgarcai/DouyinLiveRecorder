package config

import (
    "fmt"
    "strings"
)

// Service 提供配置读取与渲染的应用服务。
// 封装 Adapter，并对外提供简化接口，便于 IPC 暴露。
type Service struct{
    adapter *ConfigAdapter
}

// NewService 构造服务并加载指定根目录的配置文件。
func NewService(root string) (*Service, error) {
    a, err := LoadConfig(root)
    if err != nil { return nil, err }
    return &Service{adapter: a}, nil
}

// GetRaw 返回原始配置值。
func (s *Service) GetRaw(key string) (string, bool) {
    return s.adapter.GetRaw(key)
}

// Render 返回渲染后的配置值，支持 ${key_name} 占位符。
func (s *Service) Render(key string, ctx map[string]string) (string, error) {
    return s.adapter.RenderValue(key, ctx)
}

// SetRaw 更新原始配置值（写入磁盘并刷新适配器）。
func (s *Service) SetRaw(root, key, value string) error {
    if err := Update(root, key, value); err != nil { return err }
    // 重新加载以刷新内存视图。
    a, err := LoadConfig(root)
    if err != nil { return err }
    s.adapter = a
    return nil
}

// GetStringOrDefault 获取字符串配置，若不存在则返回默认值。
func (s *Service) GetStringOrDefault(key string, def string) string {
    if v, ok := s.adapter.GetRaw(key); ok { return v }
    return def
}

// GetIntOrDefault 获取整数配置，若解析失败则返回默认值。
func (s *Service) GetIntOrDefault(key string, def int) int {
    if v, ok := s.adapter.GetRaw(key); ok {
        var n int
        _, err := fmt.Sscanf(v, "%d", &n)
        if err == nil { return n }
    }
    return def
}

// GetBoolOrDefault 获取布尔配置（是/否/true/false），若解析失败则返回默认值。
func (s *Service) GetBoolOrDefault(key string, def bool) bool {
    if v, ok := s.adapter.GetRaw(key); ok {
        lv := strings.ToLower(strings.TrimSpace(v))
        if lv == "true" || lv == "是" || lv == "yes" { return true }
        if lv == "false" || lv == "否" || lv == "no" { return false }
    }
    return def
}
