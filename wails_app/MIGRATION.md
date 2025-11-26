# Migration Guide: Python to Wails v2

This document outlines the changes and migration steps from the original Python-based DouyinLiveRecorder to the new Go/Wails v2 version.

## Architecture Changes

| Feature | Python Version | Wails v2 Version |
|---------|----------------|------------------|
| **Backend** | Python 3.10+ | Go 1.21+ |
| **Frontend** | Console / Tkinter (limited) | Vue 3 + Vite |
| **GUI Framework** | None / Tkinter | Wails (WebView2/WebKit) |
| **Concurrency** | `threading` / `asyncio` | Go Goroutines |
| **Config Format** | `config.ini` | `config.ini` (Compatible) |

## Key Improvements

1.  **Performance**: Go provides better performance and lower resource usage compared to Python, especially for concurrent monitoring of multiple streams.
2.  **UI/UX**: A modern, web-based user interface replaces the command-line interface, making it easier to manage settings and view status.
3.  **Single Binary**: The application compiles into a single native executable (or .app bundle on macOS), simplifying distribution and installation.
4.  **Maintainability**: Strong typing in Go and component-based frontend architecture improve code maintainability.

## Configuration Migration

The `config.ini` file format remains largely compatible. You can copy your existing `config/config.ini` to the new application's config directory.

**Note**: Some internal keys might have been mapped to struct fields in Go. The application automatically handles reading and writing these values.

## Feature Parity

- [x] **Stream Recording**: Full support for ffmpeg-based recording.
- [x] **Platform Support**:
    - Douyin
    - TikTok
    - Kuaishou
    - Huya
    - Douyu
    - Bilibili
- [x] **Push Notifications**: All original channels (DingTalk, WeChat, Bark, etc.) are supported.
- [x] **Cookies & Proxy**: Full support for setting cookies and proxies per platform.
- [x] **Post-Processing**: MP4 conversion and custom scripts are implemented.

## Developer Notes

- **Spiders**: Spider logic has been ported to `internal/spider`. Each platform has its own file (e.g., `douyin.go`, `tiktok.go`).
- **Manager**: `internal/recorder/manager.go` handles the recording loop and state management.
- **Frontend**: Located in `frontend/`, built with Vue 3. Use `npm install` and `npm run dev` to work on the UI.
