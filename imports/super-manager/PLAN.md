# Super Manager 实施计划

## 1. 产品定义与边界

Super Manager 是一个**单用户、自托管、Linux 服务端**的 Codex/Claude Code 远程控制工具：

- 服务端运行在安装了 Codex、Claude Code 和项目源码的工作机器上。
- Flutter 客户端提供结构化 Agent 界面，也提供独立的原始终端界面。
- 结构化界面展示消息、工具调用、命令输出、文件修改、计划、结构化提问和错误，不复刻 TUI。
- 终端界面允许查看、输入和控制真实 PTY，可在其中运行普通 shell、Codex TUI 或 Claude Code TUI。
- 不直接调用模型 HTTP API，不接入厂商 Remote Control 云端链路，不解析 ANSI/TUI 来推断 Agent 语义。

MVP 明确不做：多人协作、公开互联网暴露、云端中继、多主机编排、附件上传、后台推送通知、任意已有 shell 的接管。

Linux 是唯一服务端验收环境。开发机上的 macOS 探针只用于研究协议，不能替代目标 Linux 上的进程、PTY、systemd 和 CLI 验收。MVP 在阶段 0 固定一种 systemd Linux 发行版和 CPU 架构。

## 2. 总体架构

```text
Flutter App
  ├── Agent UI ───── HTTP/SSE ───────────────┐
  ├── Terminal UI ── Binary WebSocket ───────┤
  └── Management ── Protobuf-defined HTTP ──┤
                                              ▼
                              Super Manager Server
                              (TypeScript modular monolith)
                                ├── AG-UI transport/replay
                                ├── Execution coordinator
                                ├── SQLite event store
                                ├── Codex Adapter ── installed Codex CLI
                                ├── Claude Adapter ─ installed Claude Code CLI
                                └── PTY Adapter ──── node-pty ── tmux
```

采用 TypeScript 服务端是因为 AG-UI 官方核心库、Codex app-server 生成类型和 `node-pty` 均可直接复用，且 Node 对 NDJSON 子进程流处理成熟。Claude 适配器只执行已安装的 CLI，不引入 Claude Agent SDK。领域层只依赖内部端口，不依赖厂商 SDK、HTTP、SQLite 或 PTY；每个 runtime 在阶段 A 后只保留一种生产适配实现，不做双轨兼容。

## 3. 协议与契约

### 3.1 AG-UI 只负责 Agent 数据面

消息、run、工具、活动、状态、推理摘要和 Interrupt 使用 AG-UI 官方类型：

- `RUN_*`
- `TEXT_MESSAGE_*`
- `TOOL_CALL_*`
- `ACTIVITY_*`
- `STATE_SNAPSHOT` / `STATE_DELTA`
- `MESSAGES_SNAPSHOT`
- `REASONING_*`
- Interrupt outcome 与 `resume`

SSE 的 `data` 始终是原样 AG-UI JSON；平台不得再定义一套平行的消息、工具、提问或状态事件。推理只展示厂商明确提供的摘要，不暴露原始思维链。

Flutter 根据 AG-UI `AgentCapabilities` 决定是否显示结构化提问、reasoning、state、multimodal 等标准控件；cancel、steer 等平台控制能力来自 control proto。MVP 固定跳过 provider 工具权限审批，不展示批准/拒绝控件。其余能力必须 capability-driven，不允许用 `provider == codex/claude` 散落分支伪造能力。

### 3.2 持久恢复是平台扩展，不冒充 AG-UI 标准能力

AG-UI 标准运行入口适合一次 HTTP POST 后流式返回事件，但没有替本项目定义可持久化的 detached-run 恢复协议。因此采用以下薄扩展：

1. 初次执行向标准 run endpoint 提交原样 `RunAgentInput`，`runId` 是幂等键。
2. 服务端先提交事件到 SQLite，再向在线订阅者发布。
3. SSE `id` 是 thread 内单调递增的平台 sequence；不写入 AG-UI event body。
4. 重连使用独立的 `GET /v1/threads/{thread_id}/events?after={sequence}`，不能重新执行 prompt。
5. sequence 已超出保留窗口时返回快照并建立新游标，而不是无限保留全量日志。

Flutter 需要一个薄 AG-UI transport wrapper；不能假定标准 `HttpAgent` 或 `Last-Event-ID` 已自动提供上述持久恢复语义。

### 3.3 非 AG-UI 控制面使用 Protobuf 单一建模

以下资源和操作在 `proto/super_manager/control/v1/` 中定义一次，并生成 TypeScript/Dart 类型：

- Runtime 与运行状态。
- 注册 Workspace 与工作目录。
- Thread 元数据和原生 session 映射状态。
- Execution 的 cancel/steer 能力与控制操作。
- TerminalSession 的创建、关闭和连接元数据。
- QuotaSnapshot 的已用比例、重置时间、观测时间和新鲜度。
- 本地配对、设备和健康状态。

AG-UI event 不包装进重复 proto；PTY 字节流也不转换为聊天事件。协议字段保持必要且语言无关，禁止 `map<string, Any>` 兜底。Buf lint、breaking check 和代码生成是合并门禁。

### 3.4 终端是独立数据面

终端通过二进制 WebSocket 传输 PTY 字节，控制帧仅负责 attach、resize、detach 和 close：

- `tmux` 持有可分离、可重连的 shell session。
- `node-pty` 只负责服务端与 tmux attach 之间的 PTY 桥接。
- API 服务重连只重建 attach，不能重启 shell。
- tmux runtime 必须处于独立 systemd cgroup/服务单元，不能作为 API service 的普通子进程，否则 API 重启会把持久终端一并杀死。
- 只能可靠接管由本系统创建并登记的 tmux session，不宣称可附着任意已有终端。
- 每个 TerminalSession 只允许一个应用级可写 attachment；慢 WebSocket 超过队列上限时只断开 attachment，不能阻塞或杀死 tmux session。
- 重连恢复当前屏幕；历史输出依赖有上限的 tmux scrollback/copy-mode，不写入 Agent event log，也不承诺逐字节永久回放。

在终端里启动的 Codex/Claude TUI 是纯终端会话：不解析屏幕、不投影为 AG-UI，也不与结构化 Agent thread 假装同步。

## 4. 正确的生命周期模型

不能把“一个厂商 turn”等同于“一个 AG-UI run”。领域层使用四个概念：

- **Thread**：长期对话，关联一个 provider session。
- **Execution**：服务端内部的一次原生 provider 执行，可跨越多个 AG-UI run。
- **Run**：一个可观察的 AG-UI 片段，以成功、错误或 interrupt 结束。
- **PendingInterrupt**：原生结构化提问句柄与 AG-UI interrupt 的持久映射。

原生 CLI 请求用户补充输入时，适配器必须声明其 continuation 类型：

- `live_callback`：原生进程和回调仍存活，例如某些 JSON-RPC user-input request。
- `persisted_resume`：原生进程已正常结束，provider 已把续接点写入 session，例如 Claude hook `defer`。

统一流程为：

1. 保存 PendingInterrupt、continuation 类型、provider session/请求映射和截止时间。
2. 持久化必要的 message/state snapshot。
3. 当前 AG-UI run 以 interrupt outcome 结束。
4. Flutter 回答全部未解决 interrupt，并用新的 `runId` 发起 resume run。
5. 服务端校验幂等键、状态、截止时间和完整答案：`live_callback` 解析原生回调；`persisted_resume` 启动新的原生进程并恢复同一 provider session。
6. 工具结果在新 run 中继续流出，并关联原始 `toolCallId`。

客户端断线不取消 Execution，但提问必须有服务端截止时间；超时后明确取消或标记 expired，不能永久占用进程。

服务端崩溃会丢失内存中的 provider callback/control request/JSON-RPC request。重启后必须把对应 `live_callback` interrupt 标记为 expired、把原生 Execution 标记为 lost，并要求用户显式重试；不能仅凭 SQLite 记录声称可继续同一次原生调用。`persisted_resume` 只有在目标 CLI 版本的契约测试确认 session 和 deferred tool call 仍存在时才可恢复；恢复会创建新的 Execution，不能伪称旧进程仍在运行。

## 5. Runtime 适配器

### 5.1 Codex：直接使用本地 app-server

公开实现检视后不再保留 `@openai/codex-sdk` 候选。官方 `codex app-server` 本来就是 Codex 富客户端使用的本地结构化接口，已经覆盖 thread/turn、流式 item、interrupt、结构化用户输入、账户和 rate-limit 等本项目需要的 surface。生产 Adapter 固定为：

```text
codex app-server --listen stdio://
```

- 只连接本机子进程的 stdio JSONL，不使用当前明确为 experimental/unsupported 的 WebSocket transport。
- 除结构化提问外，MVP 只使用默认 stable surface。当前固定候选 Codex CLI `0.144.5` 的生成 schema 将 `item/tool/requestUserInput` 标记为 experimental；由于 Ask 是硬需求，Adapter 必须在 `initialize` 时显式设置 `experimentalApi: true`，并把该 opt-in 隔离在此能力。目标 Linux 或 CLI 版本变化时重新生成 schema，标记消失后删除 opt-in，不维护双轨。
- 固定 Codex CLI 精确版本，提交该版本 `generate-ts`/`generate-json-schema` 的生成物和脱敏 fixtures；升级前执行契约回归。
- `turn/start` 每次显式传 `approvalPolicy="never"` 与该固定 schema 对应的 danger-full-access sandbox object（当前形态为 `sandboxPolicy: { type: "dangerFullAccess" }`），不能只依赖进程启动参数或用户配置。
- 阶段 A 必须验证文本、命令、文件修改、错误、真正触发的结构化用户输入、cancel、steer、resume、断线后消费、账户 rate-limit/usage、未知事件和子进程退出。
- 不再用 SDK 执行任务，也不另启第二个 app-server 补能力。

### 5.2 Claude：已安装 CLI + 隐藏 stdio control，禁止 Agent SDK

不安装、不调用 `@anthropic-ai/claude-agent-sdk`。官方 SDK quickstart 要求 `ANTHROPIC_API_KEY`，并对第三方向用户提供 claude.ai 登录/额度作出限制，这与本项目“复用目标机器上既有 Claude Code 登录、不直接调用模型 API”的边界不匹配。

公开实现扫描发现，HAPI 当前没有 Agent SDK 依赖，而是直接驱动已安装 Claude Code 的隐藏 stdio control protocol；其实现已经覆盖 `control_request/control_response`、bypass 和 `AskUserQuestion`。这与本项目对 Claude Code 2.1.195 的静态探针相互验证。因此阶段 A 的首选生产候选改为：

```text
CLAUDE_CODE_ENTRYPOINT=sdk-ts   # 仅在固定版本契约要求时设置
claude -p
  --input-format stream-json
  --output-format stream-json
  --verbose
  --include-partial-messages
  --replay-user-messages
  --permission-mode bypassPermissions
  --permission-prompt-tool stdio
  [--resume <provider-session-id>]
```

这里的 `--permission-prompt-tool stdio`、entrypoint 和 control message 是**私有、版本绑定的 SDK-host wire mode**，但本项目没有安装或调用 Agent SDK，也不注入 API key。只有目标 Linux 上的固定 CLI 版本动态验证通过后才能启用：

- prompt/输入走 stdin，不进入 argv；逐行解析 stdout，stderr 独立持续排空。
- 禁止 `--bare`：该模式跳过 OAuth/keychain，要求 API key 或 apiKeyHelper。
- 启动时执行并校验 `claude auth status --json`；必须是允许的 subscription 登录，同时显式移除 `ANTHROPIC_API_KEY`、`ANTHROPIC_AUTH_TOKEN` 等 API 凭据环境变量，否则 fail closed。
- 固定 CLI 精确版本，为 NDJSON、control request/response、session resume、未知事件和退出语义建立独立 schema、fixtures 和启动 canary；协议不匹配即拒绝启动 Structured Claude。
- `-p` 会跳过 workspace trust 对话框，因此只允许已登记、校验 realpath 的 workspace。
- 每次新执行和 `--resume` 都显式传 `bypassPermissions`；禁止 root/sudo。Bash、Edit、Write、网络和 MCP 工具在专用 Unix 账号权限内直接执行，Flutter 只展示过程与结果。

**AskUserQuestion** 使用 `can_use_tool` control request，不再依赖 TUI 或普通权限审批：

1. 对 Bash/Edit/Write/MCP 等非问题工具立即返回 `allow + 原 updatedInput`，保持 bypass 语义。
2. 对 `AskUserQuestion` 按原生 `request_id/toolUseId` 持久化 PendingInterrupt，标记为 `live_callback`，并结束当前 AG-UI run 为 interrupt；Claude 子进程继续等待 control response。
3. Flutter 提交全部结构化答案后，服务端校验幂等键、问题 schema、输入哈希、状态和截止时间，再向同一进程发送 `control_response(success, allow, updatedInput.answers)`。
4. 每个并行 control request 独立关联；普通工具可继续自动放行，问题工具分别等待。不能用“只让模型串行提问”的 prompt 代替协议正确性。
5. 超时、cancel 或客户端显式拒绝时发送对应 deny/cancel；服务端或 CLI 进程崩溃后 pending callback 必须标记为 expired/lost，不能仅凭 SQLite 伪造续接。
6. cancel 优先使用验证通过的 `interrupt` control request；若固定版本不支持，再使用受控信号，并只声明已验证的 capability。

阶段 A 必须在真实 Bash、Edit、Write、MCP、AskUserQuestion、Agent、ExitPlanMode、多个并行 tool call、断网和服务重启场景证明上述行为。HAPI 只能作为行为 oracle；未选择兼容 AGPL 的整体发布策略时，不复制其源码。

官方 `PreToolUse defer` hook 保留为一次性研究备选，而不是并行生产 fallback。它只在一个 turn 恰好一个 tool call 时可靠；若隐藏 stdio control 在目标版本失败，则必须通过 ADR 在“接受该限制并单选 hook 实现”与“本版本不提供 Structured Claude”之间选择。生产仓库不能同时维护 control 与 hook 两套执行路径。

### 5.3 Claude background supervisor：逆向研究边界

公开的 Agent View/`agents --json`、attach、logs、stop、respawn 可用于生命周期观察，但没有公开的非 TTY 结构化消息、reply 和 approval API。对本机 Claude Code 2.1.195 的已授权静态检查发现一个私有、版本绑定的 supervisor 协议：

- Unix domain socket 上的 newline-delimited JSON，带 protocol version、同 UID 校验和 control key。
- `list/has/subscribe/dispatch/attach/resize/reply/kill/respawn` 等操作可覆盖后台任务状态、原始 PTY 订阅和文本输入。
- `subscribe` 主要给出 state patch 与原始终端 stream，不是稳定的 assistant message/tool-call 语义流；禁止解析 ANSI 来伪造 AG-UI。
- 当前版本静态代码中的 `permission-response` 路径只完成鉴权和 ACK，未发现把审批决定送回 agent 的实现；必须视为不可用，不能据此承诺远程审批。

同一二进制中还存在一条与 SDK host 通信的私有 stdio control protocol，包含 `control_request/control_response` 和 pending user-dialog 状态。HAPI 的公开实现进一步证明该路线可驱动 bypass 下的 `AskUserQuestion`；它是阶段 A 的首选 Structured Claude 候选，而不是 background supervisor 的一部分。

因此 background 私有协议**不作为 Structured Claude 的事实源**。阶段 A 分别实现隔离的 `claude-bg-probe` 与 `claude-stdio-probe`：前者通过 ADR 后最多提升为可选的 background 生命周期/原始终端 helper；后者验证通过后才可成为 5.2 的唯一生产 Adapter。两者都必须固定 CLI 版本、只连接本机 IPC、协议不匹配即关闭能力，且不得把私有 roster/state/transcript 当作持久业务模型。

### 5.4 映射原则

- provider 原生 ID 只存在于 Adapter/持久映射，不泄露为客户端业务主键。
- 未知源事件保存为受限诊断样本或忽略并计数，不能导致进程崩溃。
- `RAW` 默认关闭；开启时也必须脱敏、限额、短保留，不直接展示。
- 缺失能力通过 capability 明确报告，不用 TUI 文本猜测补齐。
- 两个 runtime 追求相同的 AG-UI 语义类别，不追求虚假的功能完全一致。

## 6. 持久化、实时性与资源边界

SQLite 使用 WAL；append-only、脱敏后的 AG-UI event log 是唯一事实源：

- message、thread、execution、interrupt 和 UI state 是投影。
- event append 与投影更新在同一事务内完成。
- 只在事务提交成功后发布 SSE，保证 persist-before-publish。
- `runId` 和 interrupt resume 都有唯一约束，保证重试幂等。
- provider 原始事件只保留受控诊断环，不作为第二套业务历史。

所有资源必须有上限：

- stdout/stderr 持续排空；长输出微批为 `ACTIVITY_DELTA`。
- 达到单次输出预算后写入一次 truncation marker，随后继续排空但丢弃超额字节，避免 CLI 管道阻塞。
- 限制 frame、队列、数据库批次、run 时长、提问等待、并发和重试。
- 慢客户端只丢失自己的在线订阅，随后通过 sequence 重放。
- event log 按窗口压缩；过旧游标走 snapshot resync。
- 凭据、环境变量、cookie、Authorization、完整敏感命令输入不得写入日志、SQLite 或通知。

移动系统会暂停前台 SSE。MVP 只保证前台实时流和重新打开后的事件重放；没有独立 APNs/FCM 服务时，不承诺应用被挂起或杀死后的通知。

## 7. Workspace、并发和文件变更

- Workspace 必须预先登记；服务端对输入路径做 canonical realpath 和 symlink 边界校验。
- 每个 Workspace 同时最多一个由服务端管理的可写 Structured Execution，避免两个 Agent 竞争修改同一目录。
- 原始终端是有意提供的高权限逃生口，用户可绕过应用级 writer lease；UI 必须展示活动 writer 并警告冲突，不能把 advisory lease 宣称为安全隔离。
- 更强并行能力以后只能通过独立 Git worktree/沙箱实现，不在同一工作树上放开并发写入。

文件变更必须标明来源：

- provider 原生 Edit/Write/file-change event 只表示该 provider 明确报告的修改。
- Bash 可绕过这些事件修改任意文件，因此不能宣称原生事件是完整 diff。
- Git workspace 可在稳定边界额外采集显式 `git status/diff` 快照，并明确其采集时间。
- 非 Git workspace 或超出大小预算时，显示“无法提供完整差异”，不能猜测。

## 8. 额度与 eBPF

QuotaSnapshot 是账户级遥测，必须包含来源、观测时间和 `fresh/stale/unavailable` 状态；单轮 token、上下文窗口和估算成本另行展示，不混为套餐额度。

- Codex：只使用最终选定 Adapter 已公开的结构化能力。若选择 app-server，可验证 `account/rateLimits/read`、`account/usage/read` 和更新通知；若选定 SDK 没有等价公开接口，则显示 unavailable，不另启第二套 runtime。
- Claude：阶段 A 只记录公开 headless CLI 实际输出。若固定版本稳定发出结构化 `rate_limit_event`，可被动投影并标记观测时间；官方 CLI 契约未保证或本次会话未出现时就显示 stale/unavailable。
- 不把 Claude `statusLine` 作为产品数据源，因为它是用户唯一配置、可能改变现有终端体验，且不能保证在 headless 生命周期中执行。
- 不解析 `/usage` 展示文本，不调用私有额度端点，不做 TLS MITM。

eBPF 在 Linux 上只能作为可选诊断/研究模块：socket/tc probe 能稳定观察连接、流量和时延，但 HTTPS 正文是密文；TLS/解析函数 uprobe 依赖二进制、动态库、符号和版本。它不能成为额度功能的协议层，也不进入 MVP 关键路径。

## 9. 安全与 Linux 部署

这是远程命令执行系统；“单用户”不等于可以弱化安全：

- MVP 禁止直接暴露公网，只支持私有 overlay network 或 SSH tunnel。
- 服务端默认监听 loopback；鉴权 token 只放 header，不进入 URL。
- 配对只能由服务器本机一次性发起，生成高熵设备凭据；Flutter 存入系统 secure storage。
- 服务运行在专用 Unix 账号下，该账号单独完成 Codex/Claude 登录并只获得必要 Workspace 权限。
- Claude 启动探测必须校验 `claude auth status --json` 的认证类型，移除 API-key 类环境变量，并拒绝 `--bare`；认证详情只作内存判断，不记录账号标识。
- systemd unit 显式设置 `HOME`、`PATH`、CLI 绝对路径、工作目录和资源限制；不能假定继承交互 shell 环境。
- API service 与持久 tmux runtime 分离 cgroup；升级/重启分别验证。
- 审计创建、取消、attach、detach、提问回答、配对等控制动作，但不记录 PTY 全量按键，避免二次收集密码和 secret。
- 所有响应、结构化日志和崩溃报告统一脱敏。
- 原始终端与 Structured runtime 都具备该 Unix 账号的实际权限；MVP 主动绕过 provider permission layer，因此专用非 root 账号、Workspace 权限和网络边界才是实际安全边界。

若以后必须公网访问，需要单独设计 mTLS/device enrollment、速率限制、吊销和安全审计；不在 MVP 中用一个裸 bearer token 冒充公网安全方案。

## 10. Flutter 信息架构

### 10.1 Agent thread

```text
App Bar: Workspace / Runtime / capability / connection

Timeline
  ├── User / assistant streaming message
  ├── Plan / activity
  ├── Command card + bounded live output
  ├── Provider edit / Git snapshot diff
  ├── AskUserQuestion / structured input
  └── Error / expired / retry state

Composer: text / stop / continue
```

首版不提供附件，除非上传存储、大小限制、清理和 provider multimodal capability 均已设计完成。长日志、完整参数和诊断信息默认折叠。

### 10.2 Terminal

- 完整终端模拟器、键盘辅助栏、resize、attach/detach、连接状态。
- 明确显示 session、host、cwd 和“终端内容不会同步到 Agent timeline”。
- 网络断开不结束 tmux session，重新连接后重新 attach。

### 10.3 Dart SDK 门禁

截至本次检视，AG-UI 仓库中的 community Dart SDK 0.3.0 尚未完整覆盖新版 Interrupt lifecycle、`RunAgentInput.resume` 和 capabilities。Flutter 开发前必须：

1. 固定一个 AG-UI 协议版本。
2. 用官方 fixtures 验证 Dart JSON 编解码与 TypeScript core 一致。
3. 若上游仍缺字段，选择一个固定的最小兼容 fork 作为**唯一**依赖，只补 Interrupt、capabilities 和本项目 replay transport；不手写整套协议，不并行维护两套模型。
4. 上述门禁未通过，不开始堆叠业务页面。

## 11. 公开实现复用决策

完整扫描与源码依据见 [`docs/research/public-implementations.md`](docs/research/public-implementations.md)。结论不是“整套从零构建”，也不是“fork 一个现成产品”：

- 若可以接受 Web/PWA、AGPL 和自定义协议，直接部署 HAPI，停止自研。它是当前最接近 subscription CLI、本地 hub、terminal、bypass 与 AskUserQuestion 的公开实现。
- 若 Flutter、AG-UI、无 Claude Agent SDK 和单用户 Linux 边界不可变，没有公开项目能直接满足全部约束；继续构建薄型定制核心。
- 不整仓 fork HAPI、Happy、Paseo、Happier 或 CC Pocket。它们的 UI 栈、协议、认证或 relay 模型与目标不同，替换核心后仍会继承大量无用复杂度。
- Flutter 产品交互、问题卡、diff、配对和重连优先审计并选择性移植 MIT 的 CC Pocket 通用实现；Codex app-server 映射/fixtures 参考官方源码、Happy 与 CC Pocket；PTY/配对/资源上限参考 MIT 的 MobileCLI。
- Claude stdio control 由目标二进制探针和自有 fixtures 独立实现。未明确接受 AGPL 兼容发布策略前，不复制 HAPI/Paseo 源码。
- AG-UI、community Dart SDK、Codex 生成 schema、node-pty、tmux、SQLite、Protobuf/Buf 和成熟 Flutter terminal renderer 均直接复用，不重写协议库、VT parser 或 PTY。

仓库建立 `THIRD_PARTY_NOTICES.md`，每次移植记录来源 commit、文件、许可和修改。任何候选代码进入核心前必须通过“需求匹配、维护状态、许可、代码质量、可删除性”五项门禁。

## 12. 实施阶段与硬门禁

### 阶段 0：复用门禁与目标 Linux 基线

1. 在目标 Linux 部署固定版本 HAPI，以既有 subscription 登录完成 bypass、AskUserQuestion、cancel/resume、terminal 和断线恢复黑盒基准；若最终接受 Web/PWA、AGPL 和自定义协议，直接采用 HAPI 并终止后续自研。
2. 固定 systemd Linux 发行版、架构、Node Active LTS、tmux、Codex 和 Claude Code 精确版本。
3. 固定 AG-UI、Protobuf/Buf 及生成器版本。
4. 记录 CLI 安装、专用账号登录、Workspace 权限和 systemd 环境。
5. 建立第三方复用清单和许可门禁；只把确认需要、许可兼容且质量达标的代码带入新仓库。

### 阶段 A：协议与 Adapter Spike

所有实验必须在目标 Linux 重跑：

1. 直接验证 Codex app-server 的 stdio surface、生成 schema、ChatGPT 登录复用、bypass、cancel/steer/resume 和 rate-limit；对当前 `0.144.5` 唯一需要 experimental opt-in 的结构化用户输入单独做契约门禁，不再实现 SDK 比较分支。
2. Claude 用既有 subscription OAuth 验证 `auth status` 门禁、headless stream-json 的消息/工具/session/error/rate-limit 事件，以及 stdin 输入、cancel 和 resume。
3. 实现隔离的 `claude-stdio-probe`，验证 `permission-prompt-tool=stdio` 下 control request/response、bypass、单个与并行 AskUserQuestion、结构化答案、超时、重复回答、进程退出和服务重启后的 lost 语义。
4. 单独动态验证 background supervisor socket；默认只允许作为 terminal/lifecycle helper。若 stdio control 失败，仅用一次 `PreToolUse defer` spike 决定是否接受其单-tool-call 限制，随后通过 ADR 单选或关闭 Structured Claude。
5. 验证 AG-UI Interrupt 跨两个 run 的完整生命周期：`live_callback` 在服务重启后 expired/lost，`persisted_resume` 在契约允许时以新 Execution 恢复。
6. 验证 Dart SDK/fork 对 `outcome`、interrupt、`resume` 和 capabilities 的 JSON conformance。
7. 验证 tmux + node-pty 的 UTF-8、ANSI、resize、长输出、断网重连和 API service 重启存活。
8. 录制脱敏 provider fixtures，并建立“源事件 → AG-UI”黄金事件流。

**阶段 A 未全部通过，不进入 Flutter 业务 UI 开发。**

### 阶段 B：Contracts 与 Server

1. 建立 control proto、Buf lint/breaking/codegen。
2. 实现 Thread/Execution/Run/Interrupt 领域模型和 Workspace writer coordinator。
3. 实现选定的两个 Adapter；门禁失败的 Adapter 不留空壳兼容层。
4. 实现 AG-UI run endpoint、持久 event store、SSE replay/snapshot resync。
5. 实现 TerminalSession、PTY WebSocket 和独立 tmux runtime 管理。
6. 实现 pairing/auth、quota cache、限额、脱敏日志、健康检查和版本启动探测。

### 阶段 C：Flutter

1. 接入通过 conformance 的 AG-UI Dart 依赖和 replay transport。
2. 审计并选择性移植 CC Pocket 等 MIT 项目的通用交互组件；不得带入其 provider 消息模型、Claude SDK 假设、云推送或商业功能。
3. 实现 capability-driven timeline、Interrupt、错误和新鲜度状态。
4. 使用通过门禁的 Flutter terminal renderer 实现 Terminal UI、连接恢复和安全凭据存储，不编写 VT parser。
5. 验证 Android/iOS 前后台切换；只承诺 reopen replay，不伪造后台实时性。

### 阶段 D：Linux 部署与故障验收

1. 安装 API 与 tmux runtime 的 systemd units。
2. 通过 loopback + 私有网络/SSH tunnel 接入，不开放公网端口。
3. 验证 CLI/服务升级、API crash、provider crash、tmux crash、整机重启和磁盘满。
4. 验证 pending question 过期、重复提交、慢客户端、超量输出和旧 sequence resync。
5. 形成支持版本矩阵；未验证版本启动即失败并给出诊断，不做猜测性兼容。

## 13. 统一验收场景

对每个实际支持的 Structured runtime 执行：

1. 普通问答与长文本流。
2. 长命令实时输出、退出码和输出截断。
3. provider 原生文件修改与 Git 边界快照。
4. bypass 模式下写文件/命令不出现权限请求；AskUserQuestion 的回答、超时和重复回答正确。
5. 结构化补充输入。
6. 运行中 cancel；steer 仅在 capability 声明支持时验收。
7. Flutter 断线、Execution 继续、sequence replay。
8. Interrupt 后启动新 AG-UI run：live callback 继续同一 Execution；persisted resume 创建新 Execution 并恢复同一 provider session。
9. Interrupt 等待期间重启服务：live callback 正确显示 expired/lost；经契约验证的 persisted resume 才允许恢复。
10. CLI 未知事件、异常退出和版本不匹配。

Terminal 单独验收：shell/Codex TUI/Claude TUI、交互程序、Unicode、窗口调整、多次 detach/attach、客户端断网、API service 重启和显式关闭。

## 14. 仓库布局

```text
proto/super_manager/control/v1/  # 非 AG-UI 控制面唯一契约
server/src/domain/               # 纯领域模型
server/src/application/          # 用例与端口
server/src/adapters/codex/       # 阶段 A 后唯一 Codex 实现
server/src/adapters/claude/      # Claude CLI stream-json + 唯一选定的 control 路径
server/src/adapters/terminal/    # node-pty/tmux
server/src/transport/agui/       # AG-UI run/SSE/replay
server/src/transport/control/    # proto-defined HTTP boundary
server/src/infrastructure/       # SQLite/systemd integration/logging
app/                             # Flutter Agent UI + Terminal UI
fixtures/                        # 脱敏源事件与 AG-UI golden streams
tools/claude-bg-probe/           # 私有后台协议的隔离研究工具，不进入核心 runtime
tools/claude-stdio-probe/        # 私有 stdio control 的版本化契约探针
deploy/systemd/                  # Linux units 与安装配置
docs/adr/                        # 单选决策、版本矩阵和映射表
docs/research/                   # 公开实现、协议与复用扫描
THIRD_PARTY_NOTICES.md           # 实际移植时记录来源、commit、许可和修改
```

首个开发动作是阶段 0 与阶段 A，不是先搭 Flutter 页面，也不是先复刻任何 TUI。
