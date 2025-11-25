package legacy

import "strings"

// MapOldKeyToNew 提供旧键名到新键路径的映射，占位实现。
// 例如："录制设置.视频保存格式" -> "record.format"
func MapOldKeyToNew(old string) string {
    // 简化映射规则：中文段落替换为英文域名路径。
    s := strings.ReplaceAll(old, "录制设置", "record")
    s = strings.ReplaceAll(s, "视频保存格式", "format")
    s = strings.ReplaceAll(s, "分段时间", "segment_sec")
    s = strings.ReplaceAll(s, "质量", "quality")
    return s
}

