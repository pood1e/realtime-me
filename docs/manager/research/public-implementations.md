# 公开实现扫描与复用决策

> 快照日期：2026-07-18。该领域和 CLI 私有协议变化很快，发布前必须在目标 Linux 上重新验证固定版本。

## 1. 结论

没有一个公开项目同时满足以下全部约束：

- Flutter 原生客户端；
- AG-UI 作为唯一 Agent 数据协议；
- Linux 自托管、单用户、无公共中继；
- 复用目标机器上 Codex/Claude Code 的既有订阅登录；
- 不使用 Claude Agent SDK；
- 默认 bypass provider permission，但保留结构化 `AskUserQuestion`；
- 结构化 Agent UI 与原始 PTY 终端并存。

因此不应照搬某个现有产品，也不应从零重写所有基础设施。推荐做法是：

1. **需求可放宽为 Web/PWA + AGPL + 自定义协议时，直接部署 HAPI，停止自研。**
2. **上述约束不放宽时，保留本项目的薄型定制核心。**复用 AG-UI、Codex app-server、Flutter/PTY 组件和许可兼容的实现片段，只自建 provider 映射、持久恢复、控制面和必要 UI。
3. HAPI 作为 Claude 行为基准；CC Pocket 作为 Flutter 产品/UI 基准；Happy/CC Pocket 作为 Codex app-server 参考；MobileCLI 作为 PTY、配对和资源边界参考。

这不是“完整自研”，而是“自建不可替代的集成层”。

## 2. 候选项目

| 项目 | 许可 | 可直接复用的价值 | 与本项目的硬冲突 | 决策 |
|---|---|---|---|---|
| [HAPI](https://github.com/tiann/hapi) | AGPL-3.0-only | 本地 hub、SQLite、REST/SSE、terminal、Codex/Claude；当前 CLI 包不依赖 Claude Agent SDK；独立实现 Claude `stream-json`/stdio control，并处理 `AskUserQuestion` 与 bypass | Web/PWA，不是 Flutter；Socket.IO/自定义协议，不是 AG-UI；复制代码要求接受 AGPL 兼容的发布策略 | **最接近功能需求**。先部署做基准；需求可放宽时直接采用，否则只作行为 oracle/契约样本，不能无条件复制源码 |
| [CC Pocket](https://github.com/K9i-0/ccpocket) | MIT | Flutter + TypeScript Bridge；Codex app-server、问题卡、断线恢复、QR/Tailscale、systemd、diff/Git UI，产品形态最接近 | Claude 明确依赖 Agent SDK 与 `ANTHROPIC_API_KEY`；自定义 JSON WebSocket；无独立原始 PTY 数据面；现有 Flutter/Bridge 体量和 provider 模型较大 | **不整仓 fork**。优先抽取许可兼容、provider-agnostic 的 Flutter UI/连接模式和 Codex fixtures |
| [Happy](https://github.com/slopus/happy) | MIT | 成熟的移动/Web 远控、E2EE、重连；Codex 已迁移到 `codex app-server` stdio JSON-RPC | Claude CLI 包依赖 Agent SDK；Expo/React Native；中心 relay 和自定义 session protocol 对单用户过重 | Codex adapter、事件映射、断线测试和 E2EE 设计参考，不作产品基座 |
| [Paseo](https://github.com/getpaseo/paseo) | AGPL-3.0 | 完整 daemon、移动/桌面/Web/CLI、terminal、问题与会话恢复，产品和故障场景覆盖优秀 | Server 依赖 Claude Agent SDK；Expo；多 Agent/relay 功能远超 MVP；AGPL | UX、故障恢复和验收场景基准，不作代码基座 |
| [Happier](https://github.com/happier-dev/happier) | MIT | 多端、daemon/relay、terminal、持久 session 和丰富测试 | Claude CLI 依赖 Agent SDK；Expo/Tauri；多用户/多服务器体系过大；自定义协议 | 仅参考 daemon/relay 韧性和测试思路 |
| [MobileCLI](https://github.com/MobileCLI/mobilecli) | MIT（公开 daemon） | Rust PTY daemon、WebSocket、QR challenge-response、Tailscale、资源上限、systemd | 主要是终端镜像；通过 ANSI/输出模式识别 Agent 等待状态；移动端并非本仓库完整开放的 Flutter 基座 | PTY、配对和限流参考；禁止复用其 ANSI 语义检测路线 |
| [acpx](https://github.com/openclaw/acpx) / ACP | MIT；适配器许可各异 | 为 coding agents 提供结构化、持久 session、cancel、queue 的统一客户端协议 | 默认 Claude adapter 是 [claude-agent-acp](https://github.com/agentclientprotocol/claude-agent-acp)，其本质仍是 Claude Agent SDK；再做 ACP → AG-UI 不能消除 Claude 难点，反而多一层转换 | 暂不放入核心依赖；未来新增更多 ACP provider 时再评估 |

其他 Web UI、tmux/PTY wrapper 和云 relay 项目没有改变结论：要么解析 TUI/ANSI，要么依赖 Claude Agent SDK，要么不是 Flutter/AG-UI，要么面向多用户云服务。

## 3. 关键源码事实

### 3.1 HAPI 证明“不安装 Agent SDK也可做结构化 Claude 控制”

- [`cli/package.json`](https://github.com/tiann/hapi/blob/main/cli/package.json) 当前没有 `@anthropic-ai/claude-agent-sdk` 依赖。
- [`cli/src/claude/sdk/query.ts`](https://github.com/tiann/hapi/blob/main/cli/src/claude/sdk/query.ts) 直接启动已安装的 `claude`，使用 `--output-format stream-json`、`--input-format stream-json`、`--permission-prompt-tool stdio`，处理 `control_request/control_response`。
- [`permissionHandler.ts`](https://github.com/tiann/hapi/blob/main/cli/src/claude/utils/permissionHandler.ts) 在 `bypassPermissions` 下自动放行普通工具，但仍把 `AskUserQuestion` 作为远程输入请求，并以 `updatedInput.answers` 回送。

这与本项目对 Claude 2.1.195 的静态探针结论一致：真正值得验证的是隐藏 stdio control protocol，而不是把 background supervisor 的 PTY stream 当成结构化消息源。

但该协议仍是 Claude Code 的私有、版本绑定 surface。HAPI 的存在只能证明路线可行，不能把它变成官方兼容承诺。

### 3.2 CC Pocket 证明 Flutter 产品层无需从空白开始

[CC Pocket 的架构说明](https://k9i-0.github.io/ccpocket/architecture/)公开了 Flutter 客户端、TypeScript Bridge、Codex app-server、问题/审批、重连、离线恢复、Git/diff、QR 与 systemd 等边界。可复用的是这些交互设计、通用控件和测试场景，不是其 Claude adapter 或自定义消息模型。

其 [Bridge 文档](https://github.com/K9i-0/ccpocket/blob/main/packages/bridge/README.md)明确要求 Claude Agent SDK API key，因而不能直接满足本项目的 subscription OAuth 边界。

### 3.3 Codex 已有合适的本地结构化接口

Codex 官方 [`app-server`](https://github.com/openai/codex/blob/main/codex-rs/app-server/README.md)就是富客户端接口：本地 stdio JSONL、thread/turn 生命周期、事件、interrupt、结构化用户输入、账户和 rate-limit surface，并可从固定 CLI 版本生成 TypeScript/JSON Schema。

因此 Codex 不再保留 SDK/app-server 双候选：生产 adapter 直接选择本地 `codex app-server --listen stdio://`。

### 3.4 AG-UI 与终端基础组件可直接复用

- Agent 数据面使用官方 AG-UI TypeScript 类型；Flutter 使用 community [`ag_ui`](https://pub.dev/packages/ag_ui) 并对 Interrupt/capabilities 缺口做一次性门禁或最小 fork。
- Flutter 终端渲染优先验证 MIT 的 [`xterm.dart`](https://github.com/TerminalStudio/xterm.dart)，服务端使用 `node-pty + tmux`；不自行编写 VT parser。
- 原始 PTY 仍是独立 WebSocket 数据面，不能把 ANSI 输出反解析为 AG-UI。

## 4. 复用边界

### 可以直接依赖或移植

- AG-UI 官方 TypeScript 包、community Dart SDK；
- Codex 官方 app-server 与该固定版本生成的 schema；
- `node-pty`、tmux、SQLite、Protobuf/Buf、成熟 Flutter terminal/secure-storage 组件；
- MIT/Apache 项目中经审计的通用 UI、连接、重试、配对、fixtures 和测试场景，并保留版权/许可声明。

### 不应直接继承

- 任何并行于 AG-UI 的完整聊天/工具事件模型；
- Happy/Happier/Paseo 的多用户 relay、协作、云账户和 E2EE 存储体系；
- CC Pocket 的 Claude Agent SDK adapter 和 API-key 认证假设；
- HAPI 的 AGPL 源码，除非项目明确选择兼容 AGPL 的整体发布策略；
- MobileCLI 的 ANSI/TUI 等待状态识别。

### 必须由本项目实现

- Claude 固定版本的公开 headless/私有 stdio control 单选 adapter；
- Codex/Claude 原生事件到 AG-UI 的严格映射；
- AG-UI Interrupt 的持久记录与 provider callback 的进程内 continuation；
- 单用户控制面、事件存储、sequence replay 和 capability 投影；
- Flutter 的聚焦 Agent timeline 与独立 terminal 页面。

## 5. 目标 Linux 发布 kill switch

当前薄型核心和 Flutter MVP 已实现，但 Claude 私有协议仍必须先在目标 Linux 完成 HAPI 黑盒对照：

1. 使用既有 Claude subscription OAuth 和 Codex 登录启动真实 session。
2. 验证 bypass、`AskUserQuestion`、cancel/resume、后台运行、terminal 和断线恢复。
3. 记录它满足/不满足每个硬约束的结果。
4. 若最终可以接受 Web/PWA、AGPL 和 HAPI 自定义协议，则仍可直接采用 HAPI 并停止本项目发布。
5. 若 Flutter、AG-UI 和许可边界不变，则只保留行为 fixtures/验收结论；固定版本动态验收失败的 runtime 必须关闭，不能带风险发布。

## 6. 许可治理

- 建立 `THIRD_PARTY_NOTICES.md` 和复用清单，记录来源、commit、文件、许可和修改。
- MIT/Apache 代码只能在保留必要声明后进入仓库。
- 未明确选择 AGPL 兼容发布策略前，不复制 HAPI/Paseo 的源码；可以把部署实例作为黑盒基准。
- 许可结论在公开发布前应由项目所有者再次确认；本文只定义工程门禁，不替代法律意见。
