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
}

type RecordingSession struct {
	Url       string
	Platform  string
	StartTime time.Time
	Status    string
	StopChan  chan struct{}
}

func NewManager(ctx context.Context, cfg *config.Configuration) *Manager {
	return &Manager{
		ctx:              ctx,
		activeRecordings: make(map[string]*RecordingSession),
		config:           cfg,
		pushService:      push.NewPushService(&cfg.PushSettings),
	}
}

func (m *Manager) UpdateConfig(cfg *config.Configuration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.config = cfg
	m.pushService.UpdateConfig(&cfg.PushSettings)
}

func (m *Manager) Log(format string, args ...interface{}) {
	// Use new logger
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
		Platform:  "Douyin", // Detect platform based on URL in real impl
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

	// Default loop interval if not set
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
			// Continue monitoring
		}

		// 1. Get Stream URL
		s, err := spider.GetSpider(session.Url)
		if err != nil {
			session.Status = fmt.Sprintf("Error: %v", err)
			m.Log("Error getting spider for %s: %v", session.Url, err)
			time.Sleep(loopInterval)
			continue
		}

		// Configure Spider
		if m.config.RecordingSettings.UseProxy == "是" {
			s.SetProxy(m.config.RecordingSettings.ProxyAddress)
		}

		// Set Cookies based on domain
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
			// Stream likely offline or error
			session.Status = "Monitoring"
			// m.Log("Stream offline for %s, checking again in %v...", session.Url, loopInterval)
			time.Sleep(loopInterval)
			continue
		}

		// Stream is live!
		session.Status = "Recording"
		m.Log("Starting recording for %s from %s", session.Url, streamInfo.Url)

		// Push Start Notification
		if m.config.PushSettings.PushOnStart == "是" {
			go m.pushService.Send("Live Started", fmt.Sprintf("Recording started for %s", session.Url))
		}

		// Check Disk Space
		if err := m.checkDiskSpace(m.config.RecordingSettings.SavePath); err != nil {
			session.Status = fmt.Sprintf("Disk Error: %v", err)
			m.Log("Disk Error for %s: %v", session.Url, err)
			time.Sleep(loopInterval)
			continue
		}

		// Recording Loop (for segmentation)
		for {
			// 2. Start ffmpeg
			timestamp := time.Now().Format("20060102_150405")

			// Generate Filename
			// Format: [Date]_[Time]_[AnchorName]_[Title].[ext]
			// or user defined? For now let's stick to a standard format including Anchor and Title
			// Clean names
			anchorName := m.cleanName(streamInfo.AnchorName)
			title := m.cleanName(streamInfo.Title)
			if anchorName == "" {
				anchorName = "Unknown"
			}
			if title == "" {
				title = "Live"
			}

			// Determine extension and flags based on VideoFormat
			ext := ".ts"
			var ffmpegArgs []string
			ffmpegArgs = append(ffmpegArgs, "-i", streamInfo.Url)

			videoFormat := strings.ToLower(m.config.RecordingSettings.VideoFormat)
			if videoFormat == "mp3" || videoFormat == "m4a" || strings.Contains(videoFormat, "音频") {
				// Audio Only
				ffmpegArgs = append(ffmpegArgs, "-vn", "-c:a", "copy")
				if strings.Contains(videoFormat, "mp3") {
					ext = ".mp3"
				} else {
					ext = ".m4a"
				}
			} else {
				// Video
				ffmpegArgs = append(ffmpegArgs, "-c", "copy", "-f", "mpegts")
				if videoFormat == "mp4" {
					// Direct mp4 recording is risky for live streams (header issue), usually better to record ts then convert
					// But if user insists, we can try. However, mpegts is safer.
					// Let's stick to .ts for raw recording if format is not audio, then convert if needed.
					// Or if user selected "mp4", we might want to record as .mp4 directly?
					// The original python code often records as ts/flv first.
					// Let's keep using .ts for safety, unless it's audio.
					// Wait, if user selects "mp4", we should probably respect it or auto-convert.
					// The config has "AutoConvertToMp4".
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

			// Append output path
			ffmpegArgs = append(ffmpegArgs, outputPath)

			cmd := exec.Command("ffmpeg", ffmpegArgs...)

			if err := cmd.Start(); err != nil {
				session.Status = fmt.Sprintf("FFmpeg Error: %v", err)
				m.Log("FFmpeg Error for %s: %v", session.Url, err)
				time.Sleep(loopInterval)
				break // Break recording loop, go back to monitoring
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			// Segmentation Timer
			var splitChan <-chan time.Time
			if m.config.RecordingSettings.SplitRecording == "是" && m.config.RecordingSettings.SplitDuration > 0 {
				timer := time.NewTimer(time.Duration(m.config.RecordingSettings.SplitDuration) * time.Second)
				splitChan = timer.C
				defer timer.Stop()
			}

			// Subtitle Generation
			var subtitleStopChan chan bool
			if m.config.RecordingSettings.GenerateTimeSubtitle == "是" && ext != ".mp3" && ext != ".m4a" {
				subtitleStopChan = make(chan bool)
				go m.generateSubtitles(outputPath, subtitleStopChan)
			}

			// Wait for recording to finish, stop signal, or split timer
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
				// Stream ended naturally
				shouldStop = true
			}

			// Stop subtitle generation
			if subtitleStopChan != nil {
				subtitleStopChan <- true
				close(subtitleStopChan)
			}

			if shouldStop || shouldSplit {
				if cmd.Process != nil {
					// On Windows, Kill() is forced. On Unix, it's SIGKILL.
					// For graceful stop, we might want SIGTERM, but for stream copy, killing is usually fine for TS/FLV.
					// For MP4/M4A, it might corrupt.
					// Ideally send 'q' to stdin, but that's complex.
					// Let's try Process.Kill() for now.
					cmd.Process.Kill()
				}
			}

			// Post-processing (only if not splitting, or if we want to process every segment)
			// Usually we process every segment.
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
				// Push Stop Notification
				if m.config.PushSettings.PushOnStop == "是" {
					go m.pushService.Send("Live Stopped", fmt.Sprintf("Recording stopped for %s", session.Url))
				}
				return // Exit runRecording
			}

			if shouldSplit {
				// Continue loop to restart ffmpeg
				continue
			}

			// If we are here, it means stream ended naturally (err := <-done)
			// Break the recording loop to go back to monitoring
			break
		}

		// Push Finish Notification
		if m.config.PushSettings.PushOnStop == "是" {
			go m.pushService.Send("Live Finished", fmt.Sprintf("Recording finished for %s", session.Url))
		}

		// Loop continues to monitor...
		session.Status = "Monitoring"
		time.Sleep(loopInterval)
	}
}

func (m *Manager) GetActiveRecordings() map[string]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make(map[string]string)
	for url, session := range m.activeRecordings {
		result[url] = session.Status
	}
	return result
}

// generateSubtitles generates .srt subtitles with timestamps
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

			// Format time as HH:MM:SS,000
			startFmt := fmt.Sprintf("%02d:%02d:%02d,000", startSeconds/3600, (startSeconds%3600)/60, startSeconds%60)
			endFmt := fmt.Sprintf("%02d:%02d:%02d,000", endSeconds/3600, (endSeconds%3600)/60, endSeconds%60)

			// Write SRT entry
			entry := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", index, startFmt, endFmt, time.Now().Format("2006-01-02 15:04:05"))
			if _, err := file.WriteString(entry); err != nil {
				m.Log("Error writing subtitle: %v", err)
				return
			}
			index++
		}
	}
}

// cleanName removes emojis and special characters from string
func (m *Manager) cleanName(name string) string {
	// Remove emojis (simple regex for common ranges, might need more comprehensive one)
	// This regex matches many emoji ranges
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)
	name = emojiRegex.ReplaceAllString(name, "")

	// Remove invalid filename characters
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)
	name = invalidChars.ReplaceAllString(name, "_")

	// Trim spaces
	name = strings.TrimSpace(name)

	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// checkDiskSpace checks if there is enough disk space
func (m *Manager) checkDiskSpace(path string) error {
	if path == "" {
		path = "."
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	// Available blocks * size per block = available bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	// Threshold: 1GB
	if availableBytes < 1*1024*1024*1024 {
		return fmt.Errorf("insufficient disk space: %d MB available", availableBytes/1024/1024)
	}

	return nil
}
