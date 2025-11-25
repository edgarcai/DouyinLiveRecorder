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
		} else {
			fmt.Printf("Error loading config: %v\n", err)
			// Initialize empty config or handle error
			a.Config = &config.Configuration{}
		}
	} else {
		fmt.Println("Config file not found")
		a.Config = &config.Configuration{}
	}

	// Initialize Recorder Manager
	a.RecorderManager = recorder.NewManager(a.ctx, a.Config)
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
