package main

import (
	"context"
	"fmt"
	"wails_app/internal/config"
	"wails_app/internal/recorder"
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
