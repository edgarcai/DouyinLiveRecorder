package config

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
)

// ParseURLList 解析 URL_config.ini，返回有效URL列表（忽略注释与空行）。
func ParseURLList(root string) ([]string, error) {
    p := filepath.Join(root, "config", "URL_config.ini")
    f, err := os.Open(p)
    if err != nil { return nil, err }
    defer f.Close()
    var urls []string
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") { continue }
        if i := strings.Index(line, "#"); i >= 0 { line = strings.TrimSpace(line[:i]) }
        if line != "" { urls = append(urls, line) }
    }
    if err := sc.Err(); err != nil { return nil, err }
    return urls, nil
}

