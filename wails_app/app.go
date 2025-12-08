package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"wails_app/internal/config"
	"wails_app/internal/ffmpeg"
	"wails_app/internal/logger"
	"wails_app/internal/recorder"
	"wails_app/internal/updater"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用程序结构体
type App struct {
	ctx             context.Context
	Config          *config.Configuration
	RecorderManager *recorder.Manager
	UrlManager      *config.URLManager
	HistoryManager  *config.HistoryManager
	FFmpegManager   *ffmpeg.Manager
	UpdaterManager  *updater.Manager
}

// NewApp 创建一个新的 App 应用程序结构体
func NewApp() *App {
	return &App{}
}

// startup 在应用程序启动时调用。保存上下文
// 以便我们可以调用运行时方法
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化日志
	logger.Init(ctx)
	logger.Info(logger.LogTypeOperation, "App started")

	// 加载配置
	cfgPath := config.GetDefaultConfigPath()
	if cfgPath != "" {
		cfg, err := config.LoadConfig(cfgPath)
		if err == nil {
			a.Config = cfg
			fmt.Printf("Loaded config from %s\n", cfgPath)
			// 备份配置
			if err := config.BackupConfig(cfgPath); err != nil {
				fmt.Printf("Error backing up config: %v\n", err)
			}
		} else {
			fmt.Printf("Error loading config: %v\n", err)
			// 初始化空配置或处理错误
			a.Config = &config.Configuration{}
		}
	} else {
		fmt.Println("Config file not found")
		a.Config = &config.Configuration{}
	}

	// 初始化 URL 管理器
	// 假设配置目录与配置文件相同
	urlConfigPath := "config/URL_config.ini" // Default relative
	// TODO: 根据可执行文件位置或配置位置使用绝对路径
	a.UrlManager = config.NewURLManager(urlConfigPath)

	// 初始化历史记录管理器
	historyPath := "config/history.json"
	a.HistoryManager = config.NewHistoryManager(historyPath)

	// 初始化录制管理器
	a.RecorderManager = recorder.NewManager(a.ctx, a.Config, a.HistoryManager)

	// 初始化 FFmpeg 管理器
	a.FFmpegManager = ffmpeg.NewManager(a.ctx)

	// 初始化更新管理器
	a.UpdaterManager = updater.NewManager(a.ctx, a.Config)

	// 启动时检查更新
	if a.Config.UpdateSettings.CheckUpdateOnStart == "是" {
		go func() {
			// 给一点延迟，让前端先加载完成
			// 虽然 EventsEmit 会排队，但延迟更安全
			// 或者在前端 mounted 后发送 'app-ready' 事件再检查
			// 这里简单起见，直接检查，前端监听 update-available
			hasUpdate, _, _, _, _ := a.UpdaterManager.CheckUpdate()
			if hasUpdate {
				runtime.EventsEmit(a.ctx, "update-available")
			}
		}()
	}

	// 自动开始已持久化的 URL
	for _, url := range a.UrlManager.GetURLs() {
		go a.RecorderManager.StartRecording(url)
	}

	// 自动开始已持久化的 URL
	for _, url := range a.UrlManager.GetURLs() {
		go a.RecorderManager.StartRecording(url)
	}

	// 窗口状态恢复现在由前端在启动页后处理
}

// RestoreWindowState 恢复保存的窗口大小和位置
func (a *App) RestoreWindowState() {
	if a.Config.WindowSettings.Width > 0 && a.Config.WindowSettings.Height > 0 {
		runtime.WindowSetSize(a.ctx, a.Config.WindowSettings.Width, a.Config.WindowSettings.Height)
	} else {
		runtime.WindowSetSize(a.ctx, 1200, 800)
	}

	if a.Config.WindowSettings.X != 0 || a.Config.WindowSettings.Y != 0 {
		runtime.WindowSetPosition(a.ctx, a.Config.WindowSettings.X, a.Config.WindowSettings.Y)
	} else {
		runtime.WindowCenter(a.ctx)
	}
}

// Greet 返回给定名称的问候语
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConfig 返回当前配置
func (a *App) GetConfig() *config.Configuration {
	return a.Config
}

// UpdateConfig 更新应用程序配置
func (a *App) UpdateConfig(cfg *config.Configuration) string {
	a.Config = cfg

	// 保存到文件
	configPath := config.GetDefaultConfigPath()
	if configPath == "" {
		cwd, _ := os.Getwd()
		configPath = filepath.Join(cwd, "config", "config.ini")
	}

	err := config.SaveConfig(configPath, cfg)
	if err != nil {
		return "Error saving config: " + err.Error()
	}

	// 更新管理器
	a.RecorderManager.UpdateConfig(cfg)

	logger.Info(logger.LogTypeOperation, "Configuration updated")
	return "Saved"
}

// ToggleMiniMode 在迷你（悬浮球）和正常模式之间切换
func (a *App) ToggleMiniMode(mini bool) {
	if mini {
		// 切换到迷你模式
		logger.Info(logger.LogTypeOperation, "Switched to mini mode")
		runtime.WindowSetSize(a.ctx, 60, 60)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)

		// 如果已保存，恢复位置
		if a.Config.WindowSettings.MiniPosX != 0 || a.Config.WindowSettings.MiniPosY != 0 {
			runtime.WindowSetPosition(a.ctx, a.Config.WindowSettings.MiniPosX, a.Config.WindowSettings.MiniPosY)
		}
	} else {
		// 保存当前位置（迷你模式位置）
		x, y := runtime.WindowGetPosition(a.ctx)
		a.Config.WindowSettings.MiniPosX = x
		a.Config.WindowSettings.MiniPosY = y

		// 保存配置以持久化位置
		a.UpdateConfig(a.Config)

		// 切换到正常模式
		logger.Info(logger.LogTypeOperation, "Switched to normal mode")

		// 如果已保存，恢复正常窗口大小和位置
		width := a.Config.WindowSettings.Width
		height := a.Config.WindowSettings.Height
		if width == 0 || height == 0 {
			width = 1200
			height = 800
		}
		runtime.WindowSetSize(a.ctx, width, height)

		if a.Config.WindowSettings.X != 0 || a.Config.WindowSettings.Y != 0 {
			runtime.WindowSetPosition(a.ctx, a.Config.WindowSettings.X, a.Config.WindowSettings.Y)
		} else {
			runtime.WindowCenter(a.ctx)
		}

		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}
}

// SaveWindowState 保存当前窗口状态
func (a *App) SaveWindowState() {
	if a.ctx == nil {
		return
	}

	// 如果在迷你模式下不保存
	// 我们可以检查窗口大小来猜测模式，或跟踪状态
	// 目前，假设如果尺寸很小则是迷你模式
	w, h := runtime.WindowGetSize(a.ctx)
	if w < 200 && h < 200 {
		return
	}

	x, y := runtime.WindowGetPosition(a.ctx)
	a.Config.WindowSettings.Width = w
	a.Config.WindowSettings.Height = h
	a.Config.WindowSettings.X = x
	a.Config.WindowSettings.Y = y

	a.UpdateConfig(a.Config)
}

// StartRecording 开始录制 URL
func (a *App) StartRecording(url string) string {
	err := a.RecorderManager.StartRecording(url)
	if err != nil {
		return err.Error()
	}
	return "Started"
}

// StopRecording 停止录制 URL
func (a *App) StopRecording(url string) string {
	err := a.RecorderManager.StopRecording(url)
	if err != nil {
		return err.Error()
	}
	return "Stopped"
}

// GetRecordingStatus 返回所有活动录制的状态
func (a *App) GetRecordingStatus() []recorder.ActiveRecording {
	return a.RecorderManager.GetActiveRecordings()
}

// AddUrl 将 URL 添加到持久列表并开始录制
func (a *App) AddUrl(url string) string {
	// 更新历史记录（初始添加，尚无元数据）
	a.HistoryManager.AddOrUpdate(url, "", "", "")

	err := a.UrlManager.AddURL(url)
	if err != nil {
		return err.Error()
	}
	// 同时开始录制
	go a.RecorderManager.StartRecording(url)
	logger.Info(logger.LogTypeOperation, "Added URL: %s", url)
	return "Added"
}

// RemoveUrl 从持久列表中删除 URL 并停止录制
func (a *App) RemoveUrl(url string) string {
	err := a.UrlManager.RemoveURL(url)
	if err != nil {
		return err.Error()
	}
	// 同时停止录制
	a.RecorderManager.StopRecording(url)
	logger.Info(logger.LogTypeOperation, "Removed URL: %s", url)
	return "Removed"
}

// GetUrls 返回 URL 的持久列表
func (a *App) GetUrls() []string {
	return a.UrlManager.GetURLs()
}

// GetHistory 返回录制历史记录
func (a *App) GetHistory() []config.HistoryItem {
	return a.HistoryManager.GetHistory()
}

// RemoveHistoryItem 删除历史记录项
func (a *App) RemoveHistoryItem(url string) string {
	err := a.HistoryManager.Remove(url)
	if err != nil {
		return err.Error()
	}
	return "Removed"
}

// CheckFFmpeg 检查 ffmpeg 是否已安装
func (a *App) CheckFFmpeg() bool {
	return a.FFmpegManager.CheckFFmpeg()
}

// DownloadFFmpeg 开始下载 ffmpeg
func (a *App) DownloadFFmpeg() string {
	err := a.FFmpegManager.DownloadFFmpeg()
	if err != nil {
		return err.Error()
	}
	return "Started"
}

// GetFFmpegDownloadProgress 返回当前下载进度
func (a *App) GetFFmpegDownloadProgress() map[string]interface{} {
	progress, status := a.FFmpegManager.GetDownloadProgress()
	return map[string]interface{}{
		"progress": progress,
		"status":   status,
	}
}

// CancelFFmpegDownload 取消下载
func (a *App) CancelFFmpegDownload() {
	a.FFmpegManager.CancelDownload()
}

// GetFFmpegInfo 返回 ffmpeg 的版本信息
func (a *App) GetFFmpegInfo() string {
	return a.FFmpegManager.GetFFmpegInfo()
}

// Login 处理用户登录
func (a *App) Login(username, password string) map[string]interface{} {
	// 模拟登录逻辑
	if username != "" && password != "" {
		logger.Info(logger.LogTypeOperation, "User logged in: %s", username)
		return map[string]interface{}{
			"success":  true,
			"username": username,
			"token":    "mock_token_" + username,
		}
	}
	return map[string]interface{}{
		"success": false,
		"message": "Invalid credentials",
	}
}

// Logout 处理用户登出
func (a *App) Logout() bool {
	logger.Info(logger.LogTypeOperation, "User logged out")
	return true
}

// GetUserInfo 返回当前用户信息（模拟）
func (a *App) GetUserInfo() map[string]interface{} {
	// 在实际应用中，检查会话/令牌
	return map[string]interface{}{
		"isLoggedIn": false,
	}
}

// CheckAppUpdate 检查应用程序更新
func (a *App) CheckAppUpdate() map[string]interface{} {
	hasUpdate, newVersion, notes, downloadUrl, err := a.UpdaterManager.CheckUpdate()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	return map[string]interface{}{
		"hasUpdate":    hasUpdate,
		"newVersion":   newVersion,
		"releaseNotes": notes,
		"downloadUrl":  downloadUrl,
	}
}

// StartAppUpdate 开始更新下载过程
func (a *App) StartAppUpdate(downloadUrl string) {
	go func() {
		progressChan, filePath, err := a.UpdaterManager.DownloadUpdate(downloadUrl)
		if err != nil {
			runtime.EventsEmit(a.ctx, "update-error", err.Error())
			return
		}
		for progress := range progressChan {
			runtime.EventsEmit(a.ctx, "update-progress", progress)
		}
		runtime.EventsEmit(a.ctx, "update-complete", filePath)
	}()
}
