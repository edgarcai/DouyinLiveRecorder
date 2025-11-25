package main

import "fmt"

// RunApplication 是默认构建的占位入口。
// 在未启用 Wails 构建标签时，提供最小化可运行输出以便CI。
func RunApplication() {
    fmt.Println("DouyinLiveRecorder scaffold (no wails tag)")
}

// main 默认执行占位实现。启用 Wails 时会由 main_wails.go 接管。
func main() { RunApplication() }
