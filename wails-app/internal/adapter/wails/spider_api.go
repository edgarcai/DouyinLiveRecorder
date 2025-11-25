package wailsadapter

import (
    "douyinrecorder/pkg/app"
    "douyinrecorder/pkg/domain"
)

// SpiderService 提供房间解析与源选择的IPC接口。
type SpiderService struct{
    sp  app.SpiderProvider
    sel app.StreamSelector
}

// NewSpiderService 构造函数。
func NewSpiderService(sp app.SpiderProvider, sel app.StreamSelector) *SpiderService {
    return &SpiderService{sp: sp, sel: sel}
}

// FetchRoomInfoRPC 解析指定URL的直播间信息。
func (s *SpiderService) FetchRoomInfoRPC(url string) (domain.RoomInfo, error) {
    return s.sp.FetchRoomInfo(url)
}

// SelectSourceRPC 根据房间信息选择最佳播放源。
func (s *SpiderService) SelectSourceRPC(room domain.RoomInfo) (domain.PlaySource, error) {
    return s.sel.Select(room)
}

