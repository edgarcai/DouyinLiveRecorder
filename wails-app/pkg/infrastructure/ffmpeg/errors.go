package ffmpeg

import "strings"

// ClassifyFFmpegError 依据 stderr 文本分类错误类型，用于重试与告警。
func ClassifyFFmpegError(line string) string {
    l := strings.ToLower(line)
    switch {
    case strings.Contains(l, "connection reset") || strings.Contains(l, "network unreachable"):
        return "network"
    case strings.Contains(l, "protocol not found") || strings.Contains(l, "invalid data"):
        return "protocol"
    case strings.Contains(l, "no space left") || strings.Contains(l, "permission denied"):
        return "write"
    default:
        return "unknown"
    }
}

// ExtractSegmentPath 从 stderr 行中提取分段输出路径（匹配 Opening '...' for writing）。
func ExtractSegmentPath(line string) string {
    l := line
    i := strings.Index(l, "Opening '")
    if i < 0 { return "" }
    rest := l[i+9:]
    j := strings.Index(rest, "' for writing")
    if j < 0 { return "" }
    return rest[:j]
}
