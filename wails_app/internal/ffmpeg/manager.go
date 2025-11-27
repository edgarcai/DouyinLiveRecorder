package ffmpeg

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Manager struct {
	ctx              context.Context
	downloadCancel   context.CancelFunc
	downloadProgress float64
	downloadStatus   string // "idle", "downloading", "extracting", "completed", "error"
	mutex            sync.RWMutex
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:            ctx,
		downloadStatus: "idle",
	}
}

func (m *Manager) CheckFFmpeg() bool {
	// First check if ffmpeg is in the system path
	_, err := exec.LookPath("ffmpeg")
	if err == nil {
		return true
	}

	// Then check if it's in the current directory or a specific bin directory
	// This logic might need to be adjusted based on where we install it
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	ffmpegName := "ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
	}

	localPath := filepath.Join(cwd, ffmpegName)
	if _, err := os.Stat(localPath); err == nil {
		return true
	}

	return false
}

func (m *Manager) GetDownloadProgress() (float64, string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.downloadProgress, m.downloadStatus
}

func (m *Manager) CancelDownload() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.downloadCancel != nil {
		m.downloadCancel()
		m.downloadStatus = "cancelled"
	}
}

func (m *Manager) DownloadFFmpeg() error {
	m.mutex.Lock()
	if m.downloadStatus == "downloading" || m.downloadStatus == "extracting" {
		m.mutex.Unlock()
		return fmt.Errorf("download already in progress")
	}
	m.downloadStatus = "downloading"
	m.downloadProgress = 0
	ctx, cancel := context.WithCancel(context.Background())
	m.downloadCancel = cancel
	m.mutex.Unlock()

	defer func() {
		m.mutex.Lock()
		if m.downloadStatus != "completed" && m.downloadStatus != "cancelled" {
			m.downloadStatus = "error"
		}
		m.downloadCancel = nil
		m.mutex.Unlock()
	}()

	var url string
	var filename string

	switch runtime.GOOS {
	case "windows":
		url = "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
		filename = "ffmpeg-release-essentials.zip"
	case "darwin":
		url = "https://evermeet.cx/ffmpeg/get/zip"
		filename = "ffmpeg.zip"
	case "linux":
		// Static build for amd64, might need to detect architecture
		url = "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz"
		filename = "ffmpeg-release-amd64-static.tar.xz"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Check for existing partial download
	startByte := int64(0)
	if info, err := os.Stat(filename); err == nil {
		startByte = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength + startByte

	flags := os.O_CREATE | os.O_WRONLY
	if startByte > 0 {
		flags |= os.O_APPEND
	}

	out, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	currentBytes := startByte

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			currentBytes += int64(n)
			m.mutex.Lock()
			if totalSize > 0 {
				m.downloadProgress = float64(currentBytes) / float64(totalSize) * 100
			}
			m.mutex.Unlock()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			if err == context.Canceled {
				return err
			}
			return err
		}
	}

	m.mutex.Lock()
	m.downloadStatus = "extracting"
	m.mutex.Unlock()

	// Extract
	if err := m.extract(filename); err != nil {
		return err
	}

	// Cleanup archive
	os.Remove(filename)

	m.mutex.Lock()
	m.downloadStatus = "completed"
	m.downloadProgress = 100
	m.mutex.Unlock()

	return nil
}

func (m *Manager) extract(filename string) error {
	// Simple zip extraction for Windows/macOS
	// For Linux .tar.xz, we need a different approach, but for now let's focus on zip
	// or assume the user has tar installed on Linux

	if strings.HasSuffix(filename, ".zip") {
		return m.unzip(filename)
	} else if strings.HasSuffix(filename, ".tar.xz") {
		// Use system tar command for simplicity on Linux
		cmd := exec.Command("tar", "-xf", filename)
		return cmd.Run()
	}
	return fmt.Errorf("unsupported archive format")
}

func (m *Manager) unzip(src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// We only care about the ffmpeg binary
		if !strings.Contains(f.Name, "bin/ffmpeg") && !strings.Contains(f.Name, "ffmpeg.exe") && f.Name != "ffmpeg" {
			continue
		}

		// Flatten the path, extract directly to current directory
		fpath := filepath.Base(f.Name)

		if f.FileInfo().IsDir() {
			continue
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetFFmpegInfo() string {
	if !m.CheckFFmpeg() {
		return ""
	}

	ffmpegName := "ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegName = "ffmpeg.exe"
	}

	// Try local path first
	cwd, _ := os.Getwd()
	localPath := filepath.Join(cwd, ffmpegName)
	cmdPath := ffmpegName
	if _, err := os.Stat(localPath); err == nil {
		cmdPath = localPath
	}

	cmd := exec.Command(cmdPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error getting version: %v", err)
	}
	return string(output)
}
