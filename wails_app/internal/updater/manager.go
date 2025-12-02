package updater

import (
	"context"
	"time"
)

type Manager struct {
	ctx context.Context
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{ctx: ctx}
}

// CheckUpdate 检查是否有新版本可用
// 返回: hasUpdate, newVersion, releaseNotes, error
func (m *Manager) CheckUpdate() (bool, string, string, error) {
	// TODO: 实现真实的版本检查（例如，从 GitHub Releases 或 API 获取）
	// 目前，我们模拟检查。

	// 模拟网络延迟
	time.Sleep(1 * time.Second)

	// 模拟：将其更改为 true 以模拟更新
	// 在实际应用中，比较当前版本与远程版本
	hasUpdate := true
	newVersion := "1.1.0"
	releaseNotes := "Fix bugs and improve performance."

	return hasUpdate, newVersion, releaseNotes, nil
}

// DownloadUpdate 模拟下载。
// 返回一个发送进度 (0-100) 并在完成时关闭的通道。
func (m *Manager) DownloadUpdate() <-chan int {
	ch := make(chan int)
	go func() {
		// 模拟下载
		for i := 0; i <= 100; i++ {
			ch <- i
			// 变速以看起来逼真
			if i < 30 {
				time.Sleep(20 * time.Millisecond)
			} else if i < 70 {
				time.Sleep(50 * time.Millisecond)
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
		close(ch)
	}()
	return ch
}
