package stream

import (
    "testing"
    "douyinrecorder/pkg/domain"
)

// TestSimpleSelector 验证最小策略选择逻辑。
func TestSimpleSelector(t *testing.T) {
    s := &SimpleSelector{}
    r := domain.RoomInfo{M3U8URL: "http://m3u8", Quality: "原画"}
    ps, err := s.Select(r)
    if err != nil { t.Fatal(err) }
    if !ps.Valid || ps.URL != "http://m3u8" { t.Fatalf("unexpected: %+v", ps) }

    r2 := domain.RoomInfo{PlayURLList: []string{"http://flv"}, Quality: "高清"}
    ps2, _ := s.Select(r2)
    if !ps2.Valid || ps2.URL != "http://flv" { t.Fatalf("unexpected2: %+v", ps2) }

    r3 := domain.RoomInfo{}
    ps3, _ := s.Select(r3)
    if ps3.Valid { t.Fatalf("should be invalid") }
}

