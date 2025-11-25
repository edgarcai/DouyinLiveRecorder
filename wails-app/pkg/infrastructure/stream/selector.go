package stream

import "douyinrecorder/pkg/domain"

// SimpleSelector 提供最小策略：优先选择 M3U8URL，其次 PlayURLList 首项。
type SimpleSelector struct{}

// Select 返回最佳播放源占位实现。
func (s *SimpleSelector) Select(room domain.RoomInfo) (domain.PlaySource, error) {
    if room.M3U8URL != "" {
        return domain.PlaySource{URL: room.M3U8URL, Quality: room.Quality, Valid: true}, nil
    }
    if len(room.PlayURLList) > 0 {
        return domain.PlaySource{URL: room.PlayURLList[0], Quality: room.Quality, Valid: true}, nil
    }
    return domain.PlaySource{Valid: false}, nil
}

