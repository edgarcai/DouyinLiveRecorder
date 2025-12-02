package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建应用程序结构体的实例
	app := NewApp()

	// 使用选项创建应用程序
	err := wails.Run(&options.App{
		Title:  "DouyinLiveRecorder",
		Width:  800,
		Height: 450,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			app.SaveWindowState()
			return false
		},
		Frameless: true,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
