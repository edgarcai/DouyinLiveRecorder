package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"wails_app/internal/config"
)

type Manager struct {
	ctx    context.Context
	config *config.Configuration
}

func NewManager(ctx context.Context, cfg *config.Configuration) *Manager {
	return &Manager{ctx: ctx, config: cfg}
}

// CheckUpdate 检查是否有新版本可用
// 返回: hasUpdate, newVersion, releaseNotes, downloadUrl, error
func (m *Manager) CheckUpdate() (bool, string, string, string, error) {
	if m.config.UpdateSettings.UpdateUrl == "" {
		return false, "", "", "", fmt.Errorf("未配置更新检查地址")
	}

	// 1. 获取远程版本信息
	resp, err := http.Get(m.config.UpdateSettings.UpdateUrl)
	if err != nil {
		return false, "", "", "", fmt.Errorf("检查更新失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", "", "", fmt.Errorf("检查更新失败，服务器返回状态码: %d", resp.StatusCode)
	}

	var info struct {
		Version      string `json:"version"`
		ReleaseNotes string `json:"release_notes"`
		DownloadUrl  string `json:"download_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false, "", "", "", fmt.Errorf("解析版本信息失败: %v", err)
	}

	// 2. 比较版本
	currentVersion := "1.0.0" // TODO: 从 wails.json 或构建信息中获取
	if compareVersions(info.Version, currentVersion) > 0 {
		return true, info.Version, info.ReleaseNotes, info.DownloadUrl, nil
	}

	return false, "", "", "", nil
}

// DownloadUpdate 下载更新文件
// 返回一个发送进度 (0-100) 并在完成时关闭的通道。
func (m *Manager) DownloadUpdate(downloadUrl string) (<-chan int, string, error) {
	ch := make(chan int)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "update_*.zip") // 假设更新是 zip 包，或者根据 URL 判断
	if err != nil {
		close(ch)
		return nil, "", fmt.Errorf("创建临时文件失败: %v", err)
	}

	go func() {
		defer close(ch)
		defer tmpFile.Close()

		resp, err := http.Get(downloadUrl)
		if err != nil {
			// 实际应用中应该有更好的错误处理机制通知主线程
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return
		}

		size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

		buffer := make([]byte, 1024)
		var downloaded int

		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				tmpFile.Write(buffer[:n])
				downloaded += n
				if size > 0 {
					percentage := int(float64(downloaded) / float64(size) * 100)
					// 防止通道阻塞
					select {
					case ch <- percentage:
					default:
					}
				}
			}
			if err != nil {
				break
			}
		}
		ch <- 100
	}()

	return ch, tmpFile.Name(), nil
}

// compareVersions 比较两个版本号 v1 和 v2
// 如果 v1 > v2 返回 1, v1 < v2 返回 -1, v1 == v2 返回 0
func compareVersions(v1, v2 string) int {
	// 简单的版本比较逻辑，假设格式为 x.y.z
	// 实际项目中建议使用 go-version 等库
	return strings.Compare(v1, v2)
}
