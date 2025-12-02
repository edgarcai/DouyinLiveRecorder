package recorder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"wails_app/internal/config"
	"wails_app/internal/logger"
	"wails_app/internal/pkg/push"
	"wails_app/internal/spider"
)

type Manager struct {
	ctx              context.Context
	activeRecordings map[string]*RecordingSession
	mutex            sync.Mutex
	config           *config.Configuration
	pushService      *push.PushService
	historyManager   *config.HistoryManager
}

type RecordingSession struct {
	Url        string
	Platform   string
	StartTime  time.Time
	Status     string
	StopChan   chan struct{}
	Title      string
	AnchorName string
}

type ActiveRecording struct {
	Url        string    `json:"url"`
	Status     string    `json:"status"`
	Title      string    `json:"title"`
	AnchorName string    `json:"anchor_name"`
	Platform   string    `json:"platform"`
	StartTime  time.Time `json:"start_time"`
}

func NewManager(ctx context.Context, cfg *config.Configuration, historyManager *config.HistoryManager) *Manager {
	return &Manager{
		ctx:              ctx,
		activeRecordings: make(map[string]*RecordingSession),
		config:           cfg,
		pushService:      push.NewPushService(&cfg.PushSettings),
		historyManager:   historyManager,
	}
}

func (m *Manager) UpdateConfig(cfg *config.Configuration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.config = cfg
	m.pushService.UpdateConfig(&cfg.PushSettings)
}

func (m *Manager) Log(format string, args ...interface{}) {
	// 使用新的日志记录器
	logger.Info(logger.LogTypeRunning, format, args...)
}

func (m *Manager) StartRecording(url string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.activeRecordings[url]; exists {
		return fmt.Errorf("recording already active for %s", url)
	}

	session := &RecordingSession{
		Url:       url,
		Platform:  "Douyin", // 在实际实现中根据 URL 检测平台
		StartTime: time.Now(),
		Status:    "Starting",
		StopChan:  make(chan struct{}),
	}
	m.activeRecordings[url] = session

	go m.runRecording(session)
	return nil
}

func (m *Manager) StopRecording(url string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.activeRecordings[url]
	if !exists {
		return fmt.Errorf("no active recording for %s", url)
	}

	close(session.StopChan)
	delete(m.activeRecordings, url)
	return nil
}

func (m *Manager) runRecording(session *RecordingSession) {
	spider.InitSpiders()

	// 如果未设置，则使用默认循环间隔
	loopInterval := time.Duration(m.config.RecordingSettings.LoopInterval) * time.Second
	if loopInterval == 0 {
		loopInterval = 60 * time.Second
	}

	for {
		select {
		case <-session.StopChan:
			session.Status = "Stopped"
			m.Log("Monitoring stopped for %s", session.Url)
			return
		default:
			// 继续监控
		}

		// 1. 获取流 URL
		s, err := spider.GetSpider(session.Url)
		if err != nil {
			session.Status = fmt.Sprintf("Error: %v", err)
			m.Log("Error getting spider for %s: %v", session.Url, err)
			time.Sleep(loopInterval)
			continue
		}

		// 配置爬虫
		if m.config.RecordingSettings.UseProxy == "是" {
			s.SetProxy(m.config.RecordingSettings.ProxyAddress)
		}

		// 根据域名设置 Cookie
		if strings.Contains(session.Url, "douyin.com") {
			s.SetCookies(m.config.Cookies.Douyin)
		} else if strings.Contains(session.Url, "tiktok.com") {
			s.SetCookies(m.config.Cookies.Tiktok)
		} else if strings.Contains(session.Url, "kuaishou.com") {
			s.SetCookies(m.config.Cookies.Kuaishou)
		} else if strings.Contains(session.Url, "huya.com") {
			s.SetCookies(m.config.Cookies.Huya)
		} else if strings.Contains(session.Url, "douyu.com") {
			s.SetCookies(m.config.Cookies.Douyu)
		} else if strings.Contains(session.Url, "bilibili.com") {
			s.SetCookies(m.config.Cookies.Bilibili)
		}

		streamInfo, err := s.GetStreamUrl(session.Url)
		if err != nil {
			// 流可能离线或出错
			session.Status = "Monitoring"
			// m.Log("Stream offline for %s, checking again in %v...", session.Url, loopInterval)
			time.Sleep(loopInterval)
			continue
		}

		// 直播中！
		session.Status = "Recording"
		session.Title = streamInfo.Title
		session.AnchorName = streamInfo.AnchorName
		m.Log("Starting recording for %s from %s", session.Url, streamInfo.Url)

		// 更新带有元数据的历史记录
		if m.historyManager != nil {
			m.historyManager.AddOrUpdate(session.Url, streamInfo.Title, streamInfo.AnchorName, session.Platform)
		}

		// 推送开始通知
		if m.config.PushSettings.PushOnStart == "是" {
			go m.pushService.Send("Live Started", fmt.Sprintf("Recording started for %s", session.Url))
		}

		// 检查磁盘空间
		if err := m.checkDiskSpace(m.config.RecordingSettings.SavePath); err != nil {
			session.Status = fmt.Sprintf("Disk Error: %v", err)
			m.Log("Disk Error for %s: %v", session.Url, err)
			time.Sleep(loopInterval)
			continue
		}

		// 录制循环（用于分段）
		for {
			// 2. 启动 ffmpeg
			timestamp := time.Now().Format("20060102_150405")

			// 生成文件名
			// 格式: [日期]_[时间]_[主播名]_[标题].[扩展名]
			// 或者用户自定义？目前让我们坚持使用包含主播和标题的标准格式
			// 清理名称
			anchorName := m.cleanName(streamInfo.AnchorName)
			title := m.cleanName(streamInfo.Title)
			if anchorName == "" {
				anchorName = "Unknown"
			}
			if title == "" {
				title = "Live"
			}

			// 根据视频格式确定扩展名和标志
			ext := ".ts"
			var ffmpegArgs []string
			ffmpegArgs = append(ffmpegArgs, "-i", streamInfo.Url)

			videoFormat := strings.ToLower(m.config.RecordingSettings.VideoFormat)
			if videoFormat == "mp3" || videoFormat == "m4a" || strings.Contains(videoFormat, "音频") {
				// 仅音频
				ffmpegArgs = append(ffmpegArgs, "-vn", "-c:a", "copy")
				if strings.Contains(videoFormat, "mp3") {
					ext = ".mp3"
				} else {
					ext = ".m4a"
				}
			} else {
				// 视频
				ffmpegArgs = append(ffmpegArgs, "-c", "copy", "-f", "mpegts")
				if videoFormat == "mp4" {
					// 直接录制 mp4 对于直播流是有风险的（头文件问题），通常最好先录制 ts 然后转换
					// 但如果用户坚持，我们可以尝试。不过，mpegts 更安全。
					// 除非是音频，否则让我们坚持使用 .ts 进行原始录制，然后根据需要进行转换。
					// 或者如果用户选择了 "mp4"，我们可能希望直接录制为 .mp4？
					// 原始 python 代码通常先录制为 ts/flv。
					// 为了安全起见，让我们继续使用 .ts，除非它是音频。
					// 等等，如果用户选择 "mp4"，我们应该尊重它或自动转换。
					// 配置中有 "AutoConvertToMp4"。
					ext = ".ts"
				} else if videoFormat == "mkv" {
					ext = ".mkv"
					ffmpegArgs = []string{"-i", streamInfo.Url, "-c", "copy", "-f", "matroska"}
				} else if videoFormat == "flv" {
					ext = ".flv"
					ffmpegArgs = []string{"-i", streamInfo.Url, "-c", "copy", "-f", "flv"}
				}
			}

			filename := fmt.Sprintf("%s_%s_%s_%s%s",
				timestamp[:8], timestamp[9:], anchorName, title, ext)

			outputPath := filepath.Join(m.config.RecordingSettings.SavePath, filename)
			if m.config.RecordingSettings.SavePath == "" {
				outputPath = filename
			}

			// 追加输出路径
			ffmpegArgs = append(ffmpegArgs, outputPath)

			cmd := exec.Command("ffmpeg", ffmpegArgs...)

			if err := cmd.Start(); err != nil {
				session.Status = fmt.Sprintf("FFmpeg Error: %v", err)
				m.Log("FFmpeg Error for %s: %v", session.Url, err)
				time.Sleep(loopInterval)
				break // 中断录制循环，返回监控
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			// 分段计时器
			var splitChan <-chan time.Time
			if m.config.RecordingSettings.SplitRecording == "是" && m.config.RecordingSettings.SplitDuration > 0 {
				timer := time.NewTimer(time.Duration(m.config.RecordingSettings.SplitDuration) * time.Second)
				splitChan = timer.C
				defer timer.Stop()
			}

			// 字幕生成
			var subtitleStopChan chan bool
			if m.config.RecordingSettings.GenerateTimeSubtitle == "是" && ext != ".mp3" && ext != ".m4a" {
				subtitleStopChan = make(chan bool)
				go m.generateSubtitles(outputPath, subtitleStopChan)
			}

			// 等待录制完成、停止信号或分段计时器
			shouldStop := false
			shouldSplit := false

			select {
			case <-session.StopChan:
				m.Log("Stopping recording for %s", session.Url)
				shouldStop = true
			case <-splitChan:
				m.Log("Splitting recording for %s", session.Url)
				shouldSplit = true
			case err := <-done:
				if err != nil {
					m.Log("FFmpeg exited with error for %s: %v", session.Url, err)
				} else {
					m.Log("Recording finished for %s", session.Url)
				}
				// 流自然结束
				shouldStop = true
			}

			// 停止字幕生成
			if subtitleStopChan != nil {
				subtitleStopChan <- true
				close(subtitleStopChan)
			}

			if shouldStop || shouldSplit {
				if cmd.Process != nil {
					// 在 Windows 上，Kill() 是强制的。在 Unix 上，它是 SIGKILL。
					// 为了优雅停止，我们可能想要 SIGTERM，但对于流复制，杀死通常对 TS/FLV 是可以的。
					// 对于 MP4/M4A，它可能会损坏。
					// 理想情况下发送 'q' 到 stdin，但这很复杂。
					// 暂时尝试 Process.Kill()。
					cmd.Process.Kill()
				}
			}

			// 后处理（仅当不分段时，或者如果我们想处理每个分段）
			// 通常我们处理每个分段。
			if m.config.RecordingSettings.AutoConvertToMp4 == "是" && ext == ".ts" {
				go func(path string) {
					m.Log("Converting to MP4: %s", path)
					mp4Path := strings.Replace(path, ".ts", ".mp4", 1)
					convertCmd := exec.Command("ffmpeg", "-i", path, "-c", "copy", mp4Path)
					if err := convertCmd.Run(); err != nil {
						m.Log("Error converting to MP4: %v", err)
					} else {
						m.Log("Converted to MP4: %s", mp4Path)
						if m.config.RecordingSettings.DeleteOriginalAfterConvert == "是" {
							os.Remove(path)
						}
					}
				}(outputPath)
			}

			if m.config.RecordingSettings.RunCustomScript == "是" && m.config.RecordingSettings.CustomScriptCmd != "" {
				go func() {
					scriptCmd := exec.Command("bash", "-c", m.config.RecordingSettings.CustomScriptCmd)
					scriptCmd.Run()
				}()
			}

			if shouldStop {
				session.Status = "Stopped"
				// 推送停止通知
				if m.config.PushSettings.PushOnStop == "是" {
					go m.pushService.Send("Live Stopped", fmt.Sprintf("Recording stopped for %s", session.Url))
				}
				return // 退出 runRecording
			}

			if shouldSplit {
				// 继续循环以重启 ffmpeg
				continue
			}

			// 如果我们在这里，这意味着流自然结束 (err := <-done)
			// 中断录制循环以返回监控
			break
		}

		// 推送完成通知
		if m.config.PushSettings.PushOnStop == "是" {
			go m.pushService.Send("Live Finished", fmt.Sprintf("Recording finished for %s", session.Url))
		}

		// 循环继续监控...
		session.Status = "Monitoring"
		time.Sleep(loopInterval)
	}
}

func (m *Manager) GetActiveRecordings() []ActiveRecording {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var result []ActiveRecording
	for _, session := range m.activeRecordings {
		result = append(result, ActiveRecording{
			Url:        session.Url,
			Status:     session.Status,
			Title:      session.Title,
			AnchorName: session.AnchorName,
			Platform:   session.Platform,
			StartTime:  session.StartTime,
		})
	}
	return result
}

// generateSubtitles 生成带有时间戳的 .srt 字幕
func (m *Manager) generateSubtitles(videoPath string, stopChan chan bool) {
	srtPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".srt"
	file, err := os.Create(srtPath)
	if err != nil {
		m.Log("Error creating subtitle file: %v", err)
		return
	}
	defer file.Close()

	index := 1
	startTime := time.Now()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			startSeconds := int(elapsed.Seconds())
			endSeconds := startSeconds + 1

			// 格式化时间为 HH:MM:SS,000
			startFmt := fmt.Sprintf("%02d:%02d:%02d,000", startSeconds/3600, (startSeconds%3600)/60, startSeconds%60)
			endFmt := fmt.Sprintf("%02d:%02d:%02d,000", endSeconds/3600, (endSeconds%3600)/60, endSeconds%60)

			// 写入 SRT 条目
			entry := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", index, startFmt, endFmt, time.Now().Format("2006-01-02 15:04:05"))
			if _, err := file.WriteString(entry); err != nil {
				m.Log("Error writing subtitle: %v", err)
				return
			}
			index++
		}
	}
}

// cleanName 从字符串中删除表情符号和特殊字符
func (m *Manager) cleanName(name string) string {
	// 删除表情符号（常见范围的简单正则表达式，可能需要更全面的）
	// 此正则表达式匹配许多表情符号范围
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)
	name = emojiRegex.ReplaceAllString(name, "")

	// 删除无效的文件名字符
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// 去除空格
	name = strings.TrimSpace(name)

	// 限制长度
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// checkDiskSpace 检查是否有足够的磁盘空间
func (m *Manager) checkDiskSpace(path string) error {
	if path == "" {
		path = "."
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	// 可用块 * 每块大小 = 可用字节
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// 阈值：1GB
	if availableBytes < 1*1024*1024*1024 {
		return fmt.Errorf("insufficient disk space: %d MB available", availableBytes/1024/1024)
	}

	return nil
}
