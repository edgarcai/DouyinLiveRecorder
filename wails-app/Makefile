.PHONY: build build-win build-mac build-linux

# build 统一构建入口：若检测到 wails 命令则执行真实构建，否则给出提示。
build:
	@command -v wails >/dev/null 2>&1 && wails build || echo "wails 未安装：当前为占位构建，请安装后再执行"

build-win:
	@command -v wails >/dev/null 2>&1 && wails build -platform windows/amd64 || echo "wails 未安装：Windows 构建占位"

build-mac:
	@command -v wails >/dev/null 2>&1 && wails build -platform darwin/amd64,darwin/arm64 || echo "wails 未安装：macOS 构建占位"

build-linux:
	@command -v wails >/dev/null 2>&1 && wails build -platform linux/amd64,linux/arm64 || echo "wails 未安装：Linux 构建占位"

