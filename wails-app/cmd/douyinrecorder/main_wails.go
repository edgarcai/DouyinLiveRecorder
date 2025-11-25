//go:build wails

package main

import (
    wailsadapter "douyinrecorder/internal/adapter/wails"
    "douyinrecorder/pkg/app"
    "douyinrecorder/pkg/domain"
    "douyinrecorder/pkg/infrastructure/recorder"
    "douyinrecorder/pkg/infrastructure/spider"
    "douyinrecorder/pkg/infrastructure/stream"
    "douyinrecorder/pkg/infrastructure/js"
    f "douyinrecorder/pkg/infrastructure/ffmpeg"
    conf "douyinrecorder/pkg/infrastructure/config"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

// main（wails 构建）绑定后端服务并启动 GUI。
func main() {
    var coord app.RecorderCoordinator = recorder.NewCoordinator()
    recService := wailsadapter.NewRecorderService(coord)
    cfgService, _ := wailsadapter.NewConfigService(".")
    ds := spider.NewDouyinSpider(&js.NodeEngine{})
    ds.SetScriptPath("src/javascript/x-bogus.js")
    ds.SetUA("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
    ds.SetRetryPolicy(3, 5)
    spService := wailsadapter.NewSpiderService(ds, &stream.SimpleSelector{})
    pipeline := app.NewPipeline(ds, &stream.SimpleSelector{}, &f.Runner{})
    pipeService := wailsadapter.NewPipelineService(pipeline)
	authService, _ := wailsadapter.NewAuthService(".")

    _ = wails.Run(&options.App{
        Title:  "DouyinLiveRecorder",
        Width:  1024,
        Height: 720,
        Bind: []interface{}{recService, cfgService, spService, authService, pipeService},
        OnStartup: func(ctx *wails.Context) {
            // 注册事件到运行器，通过 Wails 事件推送至前端。
            if rc, ok := coord.(*recorder.Coordinator); ok {
                rc.RegisterEvents(
                    func(seg string) { runtime.EventsEmit(ctx, "segment_completed", seg) },
                    func(typ, msg string) { runtime.EventsEmit(ctx, "record_error", map[string]string{"type": typ, "msg": msg}) },
                )
            }
            // 从配置加载代理与UA设置（容错读取）。
            if svc, err := conf.NewService("."); err == nil {
                proxy := svc.GetStringOrDefault("录制设置.代理地址", "")
                ua := svc.GetStringOrDefault("录制设置.UA", "")
                uaListStr := svc.GetStringOrDefault("录制设置.UA列表", "")
                retries := svc.GetIntOrDefault("录制设置.重试次数", 3)
                failures := svc.GetIntOrDefault("录制设置.熔断失败次数", 5)
                apiBase := svc.GetStringOrDefault("抖音.API_URL", "")
                if proxy != "" { ds.SetProxy(proxy) }
                if ua != "" { ds.SetUA(ua) }
                if apiBase != "" { ds.SetBaseAPI(apiBase) }
                if uaListStr != "" {
                    parts := strings.Split(uaListStr, ",")
                    var trimmed []string
                    for _, p := range parts { tp := strings.TrimSpace(p); if tp != "" { trimmed = append(trimmed, tp) } }
                    if len(trimmed) > 0 { ds.SetUAList(trimmed) }
                }
                ds.SetRetryPolicy(retries, failures)
            }
        },
    })
}

// noopCoordinator 为占位实现，确保 IPC 绑定可运行。
type noopCoordinator struct{}

// StartRecord 空实现。
func (n *noopCoordinator) StartRecord(task domain.RecordTask) error { return nil }

// StopRecord 空实现。
func (n *noopCoordinator) StopRecord(taskID string) error { return nil }
