package recorder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"wails_app/internal/config"
	"wails_app/internal/pkg/push"
	"wails_app/internal/spider"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

func (m *Manager) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg) // Keep stdout
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "log", msg)
	}
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

		streamUrl, err := s.GetStreamUrl(session.Url)
		if err != nil {
			// Stream likely offline or error
			session.Status = "Monitoring"
			// m.Log("Stream offline for %s, checking again in %v...", session.Url, loopInterval)
			time.Sleep(loopInterval)
			continue
		}

		// Stream is live!
		session.Status = "Recording"
		m.Log("Starting recording for %s from %s", session.Url, streamUrl)

		// Push Start Notification
		if m.config.PushSettings.PushOnStart == "是" {
			go m.pushService.Send("Live Started", fmt.Sprintf("Recording started for %s", session.Url))
		}

		// 2. Start ffmpeg
		timestamp := time.Now().Format("20060102_150405")
		outputPath := fmt.Sprintf("%s/recording_%s.ts", m.config.RecordingSettings.SavePath, timestamp)
		if m.config.RecordingSettings.SavePath == "" {
			outputPath = fmt.Sprintf("recording_%s.ts", timestamp)
		}

		// Ensure directory exists
		// os.MkdirAll(filepath.Dir(outputPath), 0755) // TODO: Add filepath import if needed, or assume SavePath exists

		cmd := exec.Command("ffmpeg", "-i", streamUrl, "-c", "copy", "-f", "mpegts", outputPath)

		if err := cmd.Start(); err != nil {
			session.Status = fmt.Sprintf("FFmpeg Error: %v", err)
			m.Log("FFmpeg Error for %s: %v", session.Url, err)
			time.Sleep(loopInterval)
			continue
		}

		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		// Wait for recording to finish or stop signal
		select {
		case <-session.StopChan:
			m.Log("Stopping recording for %s", session.Url)
			if err := cmd.Process.Kill(); err != nil {
				m.Log("Failed to kill process: %v", err)
			}
			session.Status = "Stopped"

			// Push Stop Notification
			if m.config.PushSettings.PushOnStop == "是" {
				go m.pushService.Send("Live Stopped", fmt.Sprintf("Recording stopped for %s", session.Url))
			}
			return // Exit the loop and function

		case err := <-done:
			if err != nil {
				m.Log("FFmpeg exited with error for %s: %v", session.Url, err)
			} else {
				m.Log("Recording finished for %s", session.Url)

				// Post-processing
				if m.config.RecordingSettings.AutoConvertToMp4 == "是" {
					m.Log("Converting to MP4...")
					mp4Path := strings.Replace(outputPath, ".ts", ".mp4", 1)
					convertCmd := exec.Command("ffmpeg", "-i", outputPath, "-c", "copy", mp4Path)
					if err := convertCmd.Run(); err != nil {
						m.Log("Error converting to MP4: %v", err)
					} else {
						m.Log("Converted to MP4: %s", mp4Path)
						if m.config.RecordingSettings.DeleteOriginalAfterConvert == "是" {
							os.Remove(outputPath)
							m.Log("Deleted original file: %s", outputPath)
						}
					}
				}

				if m.config.RecordingSettings.RunCustomScript == "是" && m.config.RecordingSettings.CustomScriptCmd != "" {
					m.Log("Running custom script...")
					scriptCmd := exec.Command("bash", "-c", m.config.RecordingSettings.CustomScriptCmd)
					if err := scriptCmd.Run(); err != nil {
						m.Log("Error running custom script: %v", err)
					} else {
						m.Log("Custom script executed successfully")
					}
				}

				// Push Finish Notification
				if m.config.PushSettings.PushOnStop == "是" {
					go m.pushService.Send("Live Finished", fmt.Sprintf("Recording finished for %s", session.Url))
				}
			}
			// Loop continues to monitor...
			session.Status = "Monitoring"
			time.Sleep(loopInterval)
		}
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
