# Super Manager MVP 实施方案

## 1. 目标与边界

Super Manager 是一个单用户、自托管的 Codex/Claude Code 远程控制工具。Linux 工作机继续使用两个 CLI 已有的订阅登录；服务端不直接调用模型 HTTP API，也不引入要求 API key 的 Claude Agent SDK。

MVP 提供两条彼此独立的数据面：

- **结构化 Agent**：Flutter 展示 AG-UI 消息、工具、命令输出、文件修改、推理摘要、错误和结构化提问，不复刻 TUI。
- **原始终端**：Flutter 通过二进制 WebSocket 操作本系统创建的 tmux shell，可在其中运行普通命令或 CLI TUI；终端字符流不进入 Agent 历史。

明确不做：多人协作、公共云中继、多主机编排、附件、后台推送、任意现存 shell 接管、ANSI/TUI 语义识别和 provider 普通权限审批。CLI 始终以 bypass/never 运行，但 `AskUserQuestion` 保留。

## 2. DDNS 直连拓扑

```text
Flutter Android
  │ HTTPS / SSE / WSS
  ▼
DDNS hostname ── A/AAAA ── 住宅真实公网地址
  │
  ▼
路由器转发或 IPv6 防火墙
  │
  ├── 443  Caddy：设备 mTLS + 内层 bearer
  └── 8443 Caddy：只开放一次性 PairDevice
                    │
                    ▼
              127.0.0.1:3080
              Super Manager
```

`ddns-go` 与 Caddy 和应用运行在同一台 Linux 主机，不再经过 VPS、Cloudflare Tunnel、CDN 或 relay。DDNS 只更新 DNS，不能穿透 CGNAT；上线前必须确认至少一个地址族可从外网入站。

## 3. 已实现架构

### Linux 服务端

- TypeScript/Fastify 模块化单体，生产环境强制 loopback、HTTPS 公网 origin、非 root 和 workspace allowlist。
- Protobuf + ConnectRPC 控制面：runtime、quota、workspace、thread、execution、terminal、device。
- SQLite WAL：资源、run、pending interrupt、AG-UI append-only event、设备和额度快照。
- AG-UI SSE：先持久化后发布；thread 内单调 sequence；`after` 游标重放；15 秒 heartbeat；慢订阅者有 4 MiB 上限。
- `node-pty + tmux`：最多四个持久 shell、一个可写 attachment、二进制 Protobuf WebSocket、resize/detach/close。
- 本地 PKI：私有 CA、DDNS 域名服务证书、设备 PKCS#12、可吊销 bearer、十分钟单次配对 secret。
- `smctl doctor`、PKI 初始化/续签、一次性配对、设备查询/吊销，以及 systemd/Caddy/ddns-go 部署样例。

### Flutter Android 客户端

- Material 3、亮/暗主题、Riverpod 与 go_router。
- 扫码或粘贴配对；严格校验同域 HTTPS origin、CA 摘要、有效期和 32 字节 secret。
- 私有 CA 校验 + 客户端证书 + bearer；凭据存入 Android secure storage，禁止明文 HTTP 和坏证书绕过。
- AG-UI timeline：消息、Markdown、tool/activity/reasoning、错误、cancel、Codex steer、sequence 去重与重连。
- Interrupt 表单：单选、多选、其他值、自由输入和敏感输入；必须完整回答后才能 resume。
- xterm.dart 终端：UTF-8 流式解码、辅助键、resize、detach、关闭确认和有界重连。
- runtime/额度/设备状态与设备吊销。

### 协议复用

- Agent 事件正文只使用 `@ag-ui/core` 的标准事件，不再定义平行聊天模型。
- Flutter 使用仓库内固定的 AG-UI Dart 最小 fork，只保留协议模型和 SSE parser，并补齐 Interrupt、resume、outcome 与 capabilities。
- 平台资源只在 `proto/` 定义；TypeScript 与 Dart 均由 Buf 生成。
- PTY 字节只使用 `super_manager.terminal.v1`，不包装成 Agent 消息。

## 4. Runtime 实现

### Codex 0.144.5

- 启动本地 `codex app-server --listen stdio://`，复用 ChatGPT 登录。
- 固定版本生成的 app-server TypeScript/JSON schema 已纳入仓库。
- `turn/start` 明确发送 `approvalPolicy=never` 与 `dangerFullAccess`。
- 映射文本、tool、命令输出、文件变更、计划、安全 reasoning summary、cancel、steer 和结构化 user input。
- 额度来自 app-server 的结构化 rate-limit read/update，不抓包。

### Claude Code 2.1.195

- 直接启动已安装的 `claude --print --input-format stream-json --output-format stream-json ...`。
- 使用固定版本的隐藏 stdio control 协议处理 `can_use_tool`；普通工具自动 allow，`AskUserQuestion` 转成 AG-UI Interrupt。
- 启动前校验 `claude auth status --json` 必须是第一方 claude.ai subscription，并移除 API-key 类环境变量；禁止 `--bare`。
- 支持 stream text/tool/result、session resume、cancel control 和被动 `rate_limit_event`。
- 该 control wire 不是公开稳定 API，因此目标 Linux 上版本不匹配或登录类型不符时 fail closed，不维护第二套 hook/background fallback。

Claude background supervisor 的 socket/PTY 能力不作为结构化事实源；当前 MVP 不解析它，也不把后台终端 stream 伪装成 AG-UI。

## 5. Run、Ask 与恢复语义

- **Thread** 是长期 provider session。
- **Execution** 是一次仍存活的原生 provider 调用。
- **Run** 是一个 AG-UI 片段；结构化提问会结束当前 run，回答后以新 run 继续同一 Execution。
- 每个 Workspace 同时只有一个结构化 writer，全局最多两个 Execution。

服务端收到 Ask 时持久化问题和完整性哈希，然后发出 `RUN_FINISHED(outcome=interrupt)`。Flutter 回答当前已公布的全部 interrupt；服务端校验 parent run、interrupt 集合、问题集合和选项，再唤醒原生 callback。稍后到达的并行 Ask 会在下一 continuation run 中公布，避免改变已结束 run。

resume 的敏感答案不写入 `RUN_STARTED` 或平台 raw diagnostics；Claude 问题工具结果也只记录“已提交”。Provider 自己的本地 session 仍可能保存或复述用户输入，因此“敏感”表示客户端遮蔽和平台最小留存，不是对模型的保密通道。

应用重启不会重新执行 prompt。客户端按 sequence 重放已提交事件。服务端重启后，内存中的 provider callback 无法可靠恢复，相关 Execution/Interrupt 会被标为 `LOST`；MVP 不伪造跨进程 live continuation。

## 6. 额度与 eBPF 决策

- Codex 使用 app-server 结构化额度事件。
- Claude 仅被动消费当前 CLI 会话确实发出的 `rate_limit_event`；没有观测时显示 unavailable，超过 15 分钟显示 stale。
- 不解析 `/usage` TUI，不调用私有额度端点，不做 TLS MITM。
- eBPF 只能看到连接、字节与时延；HTTPS 正文不可见，uprobe 又与二进制/TLS 实现强绑定。因此不进入额度功能和 MVP 关键路径。

## 7. 安全与资源边界

- 公网只暴露 Caddy；应用、ddns-go 管理 UI、CLI stdio、tmux socket 和 SQLite 均留在主机内。
- 443 同时要求私有 CA 签发的设备证书与可吊销 bearer；8443 只路由 PairDevice，可在不用时关闭。
- 设备证书有效 365 天，服务证书 825 天；到期前重新配对/重新签发，不自动降低验证等级。
- 服务运行在专用非 root Unix 账号；该账号和 workspace 文件权限才是 bypass 模式下的真实安全边界。
- prompt 最大 128 KiB；结构化问题、frame、paused output、SSE backlog、terminal 数量、执行并发和诊断存储均有限额。
- raw diagnostics 只保留有界、结构化且字段值脱敏的形状信息；Authorization 和设备凭据不写日志。
- 当前 AG-UI 事件历史尚未压缩，部署方需监控 SQLite 大小并备份数据目录；在实现快照/保留策略前不能宣称无限历史适合长期运行。

## 8. DDNS 部署门禁

1. 对比路由器 WAN IPv4 与外网观察值，排除私网和 `100.64.0.0/10` CGNAT。
2. 只发布真正可入站的 A/AAAA；IPv6 需同时开放路由器和 Linux 防火墙。
3. 配置 `443 -> Linux:443`、`8443 -> Linux:8443`；运营商封端口时使用外部高端口并写入两个公开 URL。
4. 使用低但合理的 DNS TTL；地址变化会产生 updater 与递归缓存收敛窗口，不能承诺零中断。
5. 验证家庭 LAN 的 NAT loopback；不支持时用路由器 split DNS 把同一 DDNS 域名解析到内网地址，不能改用不匹配证书的 IP URL。
6. 住宅网络不可入站时，DDNS profile 必须明确停用，再单独选择 overlay VPN 或国内 relay；MVP 不并行暗跑两条路径。

具体命令见 [`deploy/README.md`](deploy/README.md)。

## 9. 剩余验收

代码级检查和 Android debug 构建在开发机执行；最终发布前仍必须在目标 systemd Linux 上完成：

1. 固定 Node/CLI/tmux/OpenSSL/Caddy/ddns-go 版本并执行 `smctl doctor`。
2. 用真实 subscription 登录分别验证普通问答、长命令、文件修改、Ask、cancel、Codex steer、session continuation 和额度。
3. 验证 Flutter 断网、前后台切换、SSE replay、DDNS 地址变化、证书吊销与配对端口关闭。
4. 验证 API crash 后 Execution=LOST、tmux shell 仍存活、整机重启、磁盘不足和慢客户端。
5. 验证家庭公网 IPv4/IPv6、端口策略、NAT loopback/split DNS，并从家庭网络之外完成全链路连接。

目标 Linux 上任一私有 CLI 契约未通过时，对应 Structured runtime 必须显示不可用；原始终端仍可独立使用。
