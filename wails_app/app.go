package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"wails_app/internal/config"
	"wails_app/internal/recorder"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	Config          *config.Configuration
	RecorderManager *recorder.Manager
	UrlManager      *config.URLManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Load config
	cfgPath := config.GetDefaultConfigPath()
	if cfgPath != "" {
		cfg, err := config.LoadConfig(cfgPath)
		if err == nil {
			a.Config = cfg
			fmt.Printf("Loaded config from %s\n", cfgPath)
			// Backup config
			if err := config.BackupConfig(cfgPath); err != nil {
				fmt.Printf("Error backing up config: %v\n", err)
			}
		} else {
			fmt.Printf("Error loading config: %v\n", err)
			// Initialize empty config or handle error
			a.Config = &config.Configuration{}
		}
	} else {
		fmt.Println("Config file not found")
		a.Config = &config.Configuration{}
	}

	// Initialize URL Manager
	// Assuming config dir is same as config file
	urlConfigPath := "config/URL_config.ini" // Default relative
	// TODO: Use absolute path based on executable location or config location
	a.UrlManager = config.NewURLManager(urlConfigPath)

	// Initialize Recorder Manager
	a.RecorderManager = recorder.NewManager(a.ctx, a.Config)

	// Auto-start persisted URLs
	for _, url := range a.UrlManager.GetURLs() {
		go a.RecorderManager.StartRecording(url)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConfig returns the current configuration
func (a *App) GetConfig() *config.Configuration {
	return a.Config
}

// UpdateConfig updates the application configuration
func (a *App) UpdateConfig(cfg *config.Configuration) string {
	a.Config = cfg

	// Save to file
	configPath := config.GetDefaultConfigPath()
	if configPath == "" {
		cwd, _ := os.Getwd()
		configPath = filepath.Join(cwd, "config", "config.ini")
	}

	err := config.SaveConfig(configPath, cfg)
	if err != nil {
		return "Error saving config: " + err.Error()
	}

	// Update manager
	a.RecorderManager.UpdateConfig(cfg)

	return "Saved"
}

// ToggleMiniMode switches between mini (floating ball) and normal mode
func (a *App) ToggleMiniMode(mini bool) {
	if mini {
		// Switch to mini mode
		runtime.WindowSetSize(a.ctx, 60, 60)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)

		// Restore position if saved
		if a.Config.WindowSettings.MiniPosX != 0 || a.Config.WindowSettings.MiniPosY != 0 {
			runtime.WindowSetPosition(a.ctx, a.Config.WindowSettings.MiniPosX, a.Config.WindowSettings.MiniPosY)
		}
	} else {
		// Save current position (mini mode position)
		x, y := runtime.WindowGetPosition(a.ctx)
		a.Config.WindowSettings.MiniPosX = x
		a.Config.WindowSettings.MiniPosY = y

		// Save config to persist position
		a.UpdateConfig(a.Config)

		// Switch to normal mode
		runtime.WindowSetSize(a.ctx, 1200, 800)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
		runtime.WindowCenter(a.ctx)
	}
}

// StartRecording starts recording a URL
func (a *App) StartRecording(url string) string {
	err := a.RecorderManager.StartRecording(url)
	if err != nil {
		return err.Error()
	}
	return "Started"
}

// StopRecording stops recording a URL
func (a *App) StopRecording(url string) string {
	err := a.RecorderManager.StopRecording(url)
	if err != nil {
		return err.Error()
	}
	return "Stopped"
}

// GetRecordingStatus returns the status of all active recordings
func (a *App) GetRecordingStatus() map[string]string {
	return a.RecorderManager.GetActiveRecordings()
}

// AddUrl adds a URL to the persistent list and starts recording
func (a *App) AddUrl(url string) string {
	err := a.UrlManager.AddURL(url)
	if err != nil {
		return err.Error()
	}
	// Also start recording
	go a.RecorderManager.StartRecording(url)
	return "Added"
}

// RemoveUrl removes a URL from the persistent list and stops recording
func (a *App) RemoveUrl(url string) string {
	err := a.UrlManager.RemoveURL(url)
	if err != nil {
		return err.Error()
	}
	// Also stop recording
	a.RecorderManager.StopRecording(url)
	return "Removed"
}

// GetUrls returns the persistent list of URLs
func (a *App) GetUrls() []string {
	return a.UrlManager.GetURLs()
}
