package config

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
)

// Update 写入 composite 键（section.key）的新值；若键存在则就地更新，否则追加到节下方。
func Update(root, composite, value string) error {
    iniPath := filepath.Join(root, "config", "config.ini")
    f, err := os.Open(iniPath)
    if err != nil { return err }
    defer f.Close()

    parts := strings.SplitN(composite, ".", 2)
    section := ""
    key := composite
    if len(parts) == 2 { section, key = parts[0], parts[1] }

    var lines []string
    sc := bufio.NewScanner(f)
    curSection := ""
    updated := false
    for sc.Scan() {
        line := sc.Text()
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
            curSection = strings.Trim(trimmed, "[]")
            lines = append(lines, line)
            continue
        }
        if section != "" && curSection == section {
            if strings.HasPrefix(trimmed, key+"=") || strings.HasPrefix(trimmed, key+" =") {
                lines = append(lines, key+"="+value)
                updated = true
                continue
            }
        }
        lines = append(lines, line)
    }
    if err := sc.Err(); err != nil { return err }

    if !updated {
        if section != "" {
            // 附加到节尾：若文件未结束于该节，则附加节与键。
            lines = append(lines, "" )
            lines = append(lines, "["+section+"]")
            lines = append(lines, key+"="+value)
        } else {
            lines = append(lines, key+"="+value)
        }
    }

    content := strings.Join(lines, "\n")
    return os.WriteFile(iniPath, []byte(content), 0o644)
}

