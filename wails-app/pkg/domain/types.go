package domain

// RoomInfo 描述直播间的核心信息（与现有Python结构对齐）。
type RoomInfo struct {
    IsLive       bool     `json:"is_live"`
    AnchorName   string   `json:"anchor_name"`
    Title        string   `json:"title"`
    Platform     string   `json:"platform"`
    PlayURLList  []string `json:"play_url_list"`
    M3U8URL      string   `json:"m3u8_url"`
    FLVURL       string   `json:"flv_url"`
    Quality      string   `json:"quality"`
}

// PlaySource 表示一个可播放源及其质量与可用性。
type PlaySource struct {
    URL      string `json:"url"`
    Quality  string `json:"quality"`
    Valid    bool   `json:"valid"`
}

// RecordTask 代表一次录制任务的元数据。
type RecordTask struct {
    Room       RoomInfo  `json:"room"`
    OutputPath string    `json:"output_path"`
    Format     string    `json:"format"`
    SegmentSec int       `json:"segment_sec"`
}

