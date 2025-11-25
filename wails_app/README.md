# DouyinLiveRecorder (Wails v2)

This is a modern, cross-platform desktop application for recording live streams from various platforms (Douyin, TikTok, Kuaishou, Huya, Douyu, Bilibili, etc.). It is built using [Wails v2](https://wails.io/) (Go backend + Vue 3 frontend).

## Features

- **Multi-Platform Support**: Record from Douyin, TikTok, Kuaishou, Huya, Douyu, Bilibili, and more.
- **Real-time Monitoring**: Automatically checks for stream availability and starts recording.
- **Modern UI**: Clean, responsive interface built with Vue 3.
- **Configuration Management**: Easy-to-use settings panel for recording parameters, cookies, and accounts.
- **Push Notifications**: Integrated support for DingTalk, WeChat, Bark, Telegram, Email, Ntfy, and PushPlus.
- **Post-Processing**: Auto-convert recordings to MP4 and execute custom scripts.
- **Multi-language**: Support for Chinese and English interfaces.

## Prerequisites

- **Go**: v1.21 or later
- **Node.js**: v16 or later
- **NPM**: v8 or later
- **FFmpeg**: Must be installed and available in your system PATH.

## Development

1. Install dependencies:
   ```bash
   cd frontend
   npm install
   ```

2. Run in development mode:
   ```bash
   wails dev
   ```

## Build

To build the application for production:

```bash
wails build
```

The executable will be generated in `build/bin`.

## Configuration

The application uses a `config.ini` file located in `config/config.ini`. You can edit this file directly or use the in-app "Config" panel.

## License

MIT
