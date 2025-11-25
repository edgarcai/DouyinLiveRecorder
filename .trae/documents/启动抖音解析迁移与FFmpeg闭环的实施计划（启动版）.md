## 启动前准备
- 代码基线：冻结当前 `wails-app` 结构与已完成的服务接口（Spider/Recorder/Config/Auth/Pipeline）作为依赖。
- 质量门禁：在 CI 中启用 `go vet` 与 `golangci-lint`（零错误/零警告或豁免说明），单测覆盖率目标 ≥85%。
- 安全基线：敏感配置由 `AuthService/ConfigService` 统一读写，日志脱敏；本地密钥采用平台安全存储。
- 性能采样：新增统一采样器记录 `p50/p95/p99`（解析、计划、启动），作为验收参考。

## 阶段A：抖音解析真实迁移（D1–D12）
### 目标
- 用 Go 重写抖音解析器并接入 `JSEngine`（优先 goja），保持与 Python `src/spider.py` 字段与行为一致，联通 `SpiderService/Pipeline`。

### 启动任务
- A1 架构对齐（D1–D2）
  - 梳理 Python 抖音解析流程：入口、HTTP 详情（UA/Cookie/Headers/HTTP2）、页面与接口解析、x-bogus 签名调用、质量映射与回退策略。
  - 输出《抖音解析器设计说明》：接口签名、状态机、错误分类与降级路径、字段映射表。
- A2 JS 引擎接入（D3–D5）
  - 引入 goja 执行 `assets/js/x-bogus.js`；设置超时与异常回退到 Node 子进程；结果缓存与版本化。
  - 单测：脚本加载与执行、超时/异常回退、缓存命中。
- A3 HTTP 客户端策略（D6–D7）
  - 统一 `http.Client`（HTTP/2、代理、重试、超时分级、Header/UA 轮换）；从 `ConfigService/AuthService` 读取 Cookie/Authorization。
  - 单测：超时/重试、代理启用、Cookie 缺失与错误码处理。
- A4 解析实现与联调（D8–D12）
  - 实现 `DouyinSpider.FetchRoomInfo(url)`：元数据提取（标题/主播/状态/质量列表）、多质量源解析（m3u8/flv）、风控识别与降级返回。
  - 联通 `SpiderService/Pipeline`；与旧版进行 A/B 比对（golden tests），字段一致性为验收准线。

### 风险与缓解
- 签名变更：脚本热更新与双引擎回退；签名版本缓存与告警。
- 风控限频：指数退避与多代理出口；UA/Headers 轮换；错误码映射降级源。
- Cookie 失效：在 GUI 提供快速更新与校验；失败场景降级展示并提示。

### 验收标准
- 功能对齐：字段无差异；质量映射与回退一致。
- 性能：解析链路平均耗时缩短 ≥30%，`p95 < 50ms`（端到端 RPC）。
- 稳定：24h 模拟抖动无崩溃与死锁；错误日志可分类追踪。

## 阶段B：FFmpeg 执行闭环（D16–D22）
### 目标
- 完成 FFmpeg 运行器的启动/停止、分段与（可选）转码、后处理脚本、事件推送（Wails 构建标签下），形成录制闭环，并在 GUI 可观测。

### 启动任务
- B1 命令参数矩阵与资源守护（D16–D17）
  - `BuildArgs` 参数矩阵：分段（`-f segment/-segment_time`）、格式（TS/MP4/MKV/FLV）、音频模式（M4A/MP3）、代理与协议白名单；磁盘空间预检测与阈值预警。
  - 单测：参数组合正确性、空间检测与异常分支。
- B2 运行器实现与事件（D18–D20）
  - 进程启动/停止、stdout/stderr 解析；结构化日志字段：`task_id/platform/quality/bitrate/segment_seq/error_code`。
  - 事件推送（构建标签下）：`record_started/segment_completed/record_stopped/error`；前端订阅展示。
  - 单/集成测：启停一致性（100 次）、分段完整性（时间误差 ≤1s、丢段率 <0.1%）。
- B3 错误监控与重试（D21–D22）
  - 错误分类（网络/协议/写盘/空间不足）；指数退避（1s→30s）、最大重试 5 次、熔断与告警路径。
  - 压测：短断网/抖动模拟，自动恢复比例 ≥95%。

### 性能与验收
- 启动时延：录制启动 `p95 < 1.5s`。
- 内存占用：长跑降低 ≥20%；丢帧率 <0.1%。
- 错误恢复： ≥95% 的短暂网络错误自动恢复。

## 共同支撑与并行事项
- 指标采样：统一采样器用于 `SpiderService/Pipeline/RecorderService`，记录 `p50/p95/p99` 与错误比。
- GUI 展示：登录页（Cookie/Authorization 管理）、主界面事件与日志流；解析/录制一键链路体验。
- CI 门禁：`go vet/golangci-lint`、覆盖率门槛、PR 检查项（注释/SOLID/测试/性能/兼容/安全）。

## 里程碑与依赖
- D1–D12：抖音解析真实迁移完成；A/B 对齐通过。
- D16–D22：FFmpeg 闭环完成；事件链路打通；验收通过。
- 并行：D13–D15 联调与采样；D21–D24 IPC 性能优化。

## 交付物
- 解析器设计与实现文档；参数矩阵与运行器实现说明；事件清单与 GUI 订阅说明。
- 单元测试与集成测试报告；性能采样报表；问题清单与优化建议。

确认后我将按照上述计划启动实施，并在各阶段结束提交验收报告与度量结果。