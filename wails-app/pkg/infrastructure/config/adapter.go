package config

import (
    "bufio"
    "errors"
    "os"
    "path/filepath"
    "strings"
)

// ConfigAdapter 提供对现有 INI 配置的兼容读取与占位渲染能力。
type ConfigAdapter struct{
    values map[string]string
}

// LoadConfig 读取现有 INI 文件（如 config/config.ini），并构建键值映射。
// 仅实现必要子集，保持无第三方依赖，便于后续替换为成熟库。
func LoadConfig(root string) (*ConfigAdapter, error) {
    iniPath := filepath.Join(root, "config", "config.ini")
    f, err := os.Open(iniPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    vals := make(map[string]string)
    scanner := bufio.NewScanner(f)
    var section string
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }
        if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
            section = strings.Trim(line, "[]")
            continue
        }
        kv := strings.SplitN(line, "=", 2)
        if len(kv) != 2 { continue }
        key := strings.TrimSpace(kv[0])
        val := strings.TrimSpace(kv[1])
        // 兼容 INI 中的百分号转义（以 Python 逻辑为参考，这里保持值原样）。
        composite := key
        if section != "" {
            composite = section+"."+key
        }
        vals[composite] = val
    }
    if err := scanner.Err(); err != nil { return nil, err }
    return &ConfigAdapter{values: vals}, nil
}

// RenderValue 对包含占位符（如 ${key_name}）的值进行安全渲染。
// ctx 提供占位符键值；若缺失则保留原值以保证向下兼容。
func (c *ConfigAdapter) RenderValue(key string, ctx map[string]string) (string, error) {
    v, ok := c.values[key]
    if !ok { return "", errors.New("key not found") }
    out := v
    for {
        start := strings.Index(out, "${")
        if start < 0 { break }
        end := strings.Index(out[start:], "}")
        if end < 0 { break }
        end = start + end
        ph := out[start+2:end]
        repl, exists := ctx[ph]
        if !exists {
            // 向下兼容：保留原占位写法，不报错。
            break
        }
        out = out[:start] + repl + out[end+1:]
    }
    return out, nil
}

// GetRaw 返回原始未渲染值。
func (c *ConfigAdapter) GetRaw(key string) (string, bool) {
    v, ok := c.values[key]
    return v, ok
}

