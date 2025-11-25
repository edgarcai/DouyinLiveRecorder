package spider

import "testing"

// TestDouyinSpiderPlaceholder 验证占位解析返回的基本结构。
func TestDouyinSpiderPlaceholder(t *testing.T) {
    d := &DouyinSpider{}
    info, err := d.FetchRoomInfo("https://v.douyin.com/xxxx")
    if err != nil { t.Fatalf("error: %v", err) }
    if !info.IsLive || info.Platform != "douyin" { t.Fatalf("unexpected info: %+v", info) }
    if info.M3U8URL == "" && len(info.PlayURLList) == 0 { t.Fatalf("missing play url") }
}

