## 抖音解析迁移（优先）
- 解析器重写：在 `pkg/infrastructure/spider` 新增 `douyin` 解析实现，接口遵循 `SpiderProvider.FetchRoomInfo(url) (RoomInfo, error)`。
- 字段对齐：保证返回结构与 Python `src/spider.py` 一致（`is_live/anchor_name/title/platform/play_url_list/m3u8_url/flv_url/quality`）。
- JS 签名引擎：抽象 `JSEngine`，默认使用 `goja` 加载 `assets/js/x-bogus.js`；保留 Node 子进程后备（超时/异常回退）。
- Cookie/UA/Headers：继承 Python 抖音访问策略，统一在 `HTTP` 客户端（支持 HTTP/2）中配置；从配置服务读取账号/Authorization/Cookie。
- 容错与风控：增加重试/指数退避与风控错误识别（如验证码/限频），返回可降级源并记录指标。
- 接入链路：通过 `SpiderService` 暴露 RPC，并在 `Pipeline` 中串联解析→选择→参数构造。

## FFmpeg 执行闭环
- 运行器扩展：在 `FFmpegRunner` 中实现命令构造、进程启动/停止、stdout/stderr 日志抓取、错误分类与重试策略。
- 分段/转码/后处理：支持 `-f segment/-segment_time` 分段；可选转码（视频/音频）；完成后执行用户脚本（从配置读取）。
- 事件推送：在 Wails 构建下通过 `runtime.EventsEmit` 推送 `record_started/segment_completed/record_stopped/error`；前端订阅并展示。
- 资源与代理：支持代理注入与协议白名单；启动前磁盘空间检查，录制中断时安全退出并回收资源。

## 登录 GUI 第一阶段
- 界面内容：账号/Cookie/Authorization 管理页（读取、遮蔽显示、更新与校验）；错误提示与保存确认。
- 后端桥接：使用已实现的 `AuthService.GetKeyRPC/SetKeyRPC/GetKeyMaskedRPC` 与 `ConfigService`。
- 安全与合规：敏感值在前端仅遮蔽展示；后端日志脱敏；本地密钥存储集成（Keychain/DPAPI/Secret Service）。
- 流程化迁移：按“登录→主界面→业务模块”逐屏替换，保留兼容模式切换（legacy 子模块）。

## IPC 性能优化（p95 < 50ms）
- 采样与埋点：在 `SpiderService/ConfigService/PipelineService/RecorderService` 方法入口/出口记录耗时（`p50/p95/p99`）。
- 调优策略：复用 `http.Transport` 与连接池、减少 JSON 负载、批量/差分更新、避免频繁重建服务实例。
- 压力点分析：抓取与签名执行、参数构造、事件推送路径分别采样并设阈值告警。

## 测试扩充与压测
- 单元测试：为解析器、选择器、运行器、管线、配置/写入、登录桥接编写系统化用例；目标 ≥300，用例覆盖率 ≥85%。
- 集成测试：端到端（解析→选择→录制→后处理→推送）模拟；代理/网络抖动与异常覆盖。
- JMeter 压测：对 RPC 接口与事件推送做并发与持久压测，采样延迟分布与错误率。
- 稳定性 72h：在 CI 或专用环境运行持续录制与轮换任务，记录内存/句柄/死锁/泄漏指标，确保无异常。

## 兼容与配置迁移
- 保留 INI 文件路径与键名语义；`ConfigAdapter` 支持 `${key_name}` 渲染；`ConfigService` 写入刷新。
- 兼容报告：构建时输出键名映射与差异；渲染失败回退旧值并告警，不中断主流程。

## 安全与质量保障
- OWASP ASVS Level 2：输入校验、会话/密钥管理、依赖扫描、日志与错误处理策略。
- 资源管理：FFmpeg 进程生命周期、文件句柄与磁盘空间保护；异常时安全退出。

## 交付与版本
- 版本对应：旧 `vX.Y` → 新 `v(X+1).0`；构建清单包括 Windows/Linux/macOS 二进制与 DEB/RPM/DMG 安装包。
- 文档输出：API 迁移指南与性能优化白皮书；测试报告含 JMeter 压测与 72h 稳定性结论。

## 里程碑与实施序
1) 抖音解析真实迁移与 `JSEngine` 接入
2) FFmpeg 闭环与事件推送
3) 登录 GUI 第一阶段与主界面接入
4) IPC 性能采样与优化（达成 `p95 < 50ms`）
5) 测试扩充、JMeter 压测与 72h 稳定性

如无异议，我将按上述顺序开始实施（默认前端栈为 Vue 3），并在实现中持续保持接口/数据结构与旧项目一致，保证向下兼容。