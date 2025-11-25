## 项目目标与范围

* 将现有 Python/JS 录制器重构为 Wails v2.6.0+ 混合应用：前端 Vue 3（可选 React 18），后端 Go 1.21+。

* 完全保留现有数据结构与接口语义；对配置项（含占位符如 `${key_name}`）提供兼容适配与向下兼容方案。

* 通过 Wails IPC 提供前后端 RPC，单次调用端到端延迟 <50ms；实现跨平台打包（Windows/Linux/macOS）与安装包（DEB/RPM/DMG）。

## 当前工程评估摘要

* 现状：Python 为主（抓取/录制/编排），JS 用于站点签名（通过 PyExecJS 调用 Node），核心 CLI 入口 `main.py`，配置为 INI（`config/config.ini`, `config/URL_config.ini`），强依赖 `ffmpeg` 与系统 Node。

* 无系统化单元测试目录；以 `demo.py` 为功能验证；大量全局状态与线程并发，文件系统副作用明显。

* 迁移敏感点：PATH 注入、外部依赖安装、副作用性 INI 写入、JS 签名执行、线程共享状态、配置与日志目录结构。

## 技术架构设计

* 前端：默认选择 `Vue 3 + Vite + TypeScript`（若团队偏好 React 18，可切换 `-t react`，其余方案不变）。

* 后端：`Go 1.21+`，分层与 SOLID：

  * `pkg/domain` 领域模型与接口（房间信息、播放源、录制任务等）。

  * `pkg/app` 用例服务（抓取、源选择、录制编排、消息推送）。

  * `pkg/infrastructure` 基础设施（HTTP 客户端、FFmpeg 调用、JS 签名引擎、配置与日志）。

  * `internal/adapter/wails` 前后端绑定（IPC/RPC），`internal/adapter/legacy` 兼容层（与旧配置/API 对齐）。

* IPC：Wails `runtime.Events` + `Bind(services...)` 暴露后端方法；前端通过 `wailsjs/go/...` 调用；请求体最小化与并发复用，确保 <50ms。

* JS 签名：优先使用 `goja` 运行内嵌 JS（减少外部 Node 依赖）；必要时通过可选外部 Node 子进程适配器（可开关）。

* FFmpeg：使用 `exec.CommandContext` 管理进程与 I/O；统一参数策略与平台差异（Windows/Unix）；支持分段、转码、字幕/后处理脚本。

## 目录结构与仓库组织

* 新仓库：`wails-app/`（Wails 工程根）。

* 旧项目物理隔离：以 Git 子模块形式引入当前 Python 工程到 `legacy/python/`。

* 目录建议：

  * `frontend/`（Vue 3/React）

  * `cmd/douyinrecorder/`（GUI 主入口，可支持 headless）

  * `pkg/domain|app|infrastructure/`

  * `internal/adapter/wails|legacy/`

  * `configs/`（默认 INI + 映射表）

  * `scripts/`（打包与测试工具）

## 前后端通信方案

* 绑定服务示例（概念）：`RecorderService`、`SpiderService`、`ConfigService`、`PushService`。

* 模型传输：结构体序列化为 JSON，字段保持与现有 Python 结构一致；大对象拆分分页/分段传输。

* 性能策略：

  * 复用后端服务实例（单例），避免每次重建。

  * 前端调用批量化与差分更新（避免全量刷新）。

  * 减少深层嵌套 JSON 与无用字段。

  * 在关键 RPC 增加指标采样（`p99`）与日志埋点。

## Go 重写策略（核心业务）

* 抓取层：以 `http.Client`（含 HTTP/2 支持）与可插拔站点解析器实现；将现有 `src/spider.py` 平台函数逐步平滑迁移为 `SpiderProvider` 接口实现，优先迁移抖音/TikTok/快手/虎牙/斗鱼/哔哩哔哩等主站点。

* 源选择：复刻 `src/stream.py` 的质量映射与可用性检测为 `StreamSelector`。

* 录制编排：从 `main.py` 拆分为 `RecorderCoordinator`（任务管理）+ `FFmpegRunner`（进程）+ `PushNotifier`（事件推送）；消除全局状态，使用 `context` 与 `sync` 控制并发。

* 配置与日志：`gopkg.in/ini.v1`（或自研 INI 解析）读取现有 INI；日志采用 `zap/logrus`（二选一）统一结构化日志。

* JS 签名：将 `src/javascript/*` 迁移为 `assets/js` 并通过 `goja` 执行；对少量必须使用外部 Node 的脚本保留可插拔适配。

## 配置兼容与 `${key_name}` 迁移

* 保留现有 `config/config.ini` 与 `config/URL_config.ini` 路径与语义。

* 引入 `ConfigAdapter`：

  * 读取旧键名并映射到新领域模型；对包含模板占位的键值（如 `${key_name}`）在加载阶段执行安全渲染（Go `text/template` 或自定义渲染器）。

  * 写入时保留旧键名与格式（百分号转义规则沿用），并在日志中记录兼容与弃用提示。

  * 提供“兼容检查”命令输出映射表与差异报告。

* 回退策略：渲染失败或键缺失时，回退到旧键读取逻辑并警告，不中断主流程。

## 分阶段迁移计划

1. 登录模块（账号/Cookie/Authorization 管理）

   * 前端登录界面 + 后端 `AuthService`，先复用旧配置与 Cookie 写入，逐步引入安全存储。
2. 主界面

   * 录制任务列表、源选择、日志与推送状态；替换原 CLI 打印为 GUI 视图与事件总线。
3. 业务功能模块

   * 按平台迁移解析器与录制编排；在每次迁移完成后开启 A/B 比对与回归。

## 开发流程与自动化

* 初始化：`wails init -n douyinrecorder -t vue`（或 `-t react`）。

* 单元测试：后端 `go test` + 前端 `vitest/jest`，每日构建通过率 ≥90%。

* 覆盖率：目标 300+ 测试用例，核心路径（解析/选择/录制/推送）覆盖率 ≥85%。

* 构建：`make build` 封装 `wails build` 多平台产物；

  * Windows：NSIS 安装包（EXE/MSI）。

  * macOS：`.app` + DMG（可用 `create-dmg` 或 `wails` 自带流程）。

  * Linux：DEB/RPM（`nfpm/fpm` 集成）。

* CI/CD：GitHub Actions 触发每日构建、测试与打包；工件存储与发布版本号规则 `vX.Y → v(X+1).0`。

## 质量与性能保障

* 性能：接口响应时间缩短 30%，内存占用降低 20%。

  * 采样指标：IPC 延迟、解析耗时、FFmpeg 启动与写入速率、内存剖析（`pprof`）。

  * 优化：连接池、复用 `http.Transport`、零拷贝流 I/O、并发批处理。

* 测试：

  * 功能用例 ≥300，覆盖主要平台场景、异常与网络抖动。

  * 压测：JMeter 针对后端服务方法与 IPC 通道进行并发与长时稳定性测试；72 小时稳定性跑。

* 安全：OWASP ASVS Level 2

  * 配置与密钥安全存储（macOS Keychain/Windows DPAPI/Linux Secret Service），最小权限、输入校验、日志脱敏、更新机制与依赖扫描。

## 向下兼容策略

* API：保持函数签名与数据结构字段一致；通过 `legacy` 适配层提供旧接口名映射。

* 行为：保留原分段录制、转码、后处理脚本、代理与质量策略；新参数仅作为可选增强。

* 回滚：在 GUI 中提供“兼容模式（使用 legacy 子模块）”切换，问题时可回退。

## 交付物与版本策略

* 可执行：Windows/Linux/macOS 三平台二进制。

* 安装包：DEB/RPM/DMG。

* 文档：架构设计说明、API 迁移指南、性能优化白皮书。

* 测试报告：包含 JMeter 压测结果与 72 小时稳定性测试结论。

* 版本：旧 `vX.Y` 对应新 `v(X+1).0`；提供映射与变更日志。

## 风险与缓解

* JS 签名依赖：优先 `goja`，保留外部 Node 适配；签名脚本统一管理与缓存。

* FFmpeg 差异：按平台校验版本与参数；提供内置检查与自动修复建议。

* 配置副作用：移除运行期写入副作用，采用事务式更新与备份。

* 并发与状态：以 `context` + 明确生命周期管理替代全局共享；引入任务仓库与事件总线。

## 验证指标与验收

* IPC 单次调用 `p95 < 50ms`，端到端界面交互平均响应缩短 ≥30%。

* 内存占用降低 ≥20%，长跑 72 小时无泄漏与死锁。

* 单元测试每日通过率 ≥90%，总用例数 ≥300；压测报告与安全审计通过。

## 备注

* 默认选用 Vue 3；若需 React 18，保持架构与分层一致，仅替换前端模板与组件库。

* 后续实现将按函数级注释与 SOLID 原则编写，保证可维护性与扩展性。

