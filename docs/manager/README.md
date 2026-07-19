# Realtime Me Manager

单用户、自托管的 Codex 与 Claude Code 远程控制器。Linux 工作机运行服务和已登录的 CLI，
统一 Flutter 手机应用中的 Agent/Terminal 模块只通过可信局域网或 OpenVPN 到达；不调用模型
API，不经过 VPS/公共 relay，也不公开 Manager TCP 管理面。

## 数据面

- **Agent UI**：AG-UI over SSE，支持持久重放、工具/活动、结构化 Ask、cancel 和 Codex steer。
- **控制面**：Protobuf + ConnectRPC 管理 workspace、thread、runtime、terminal 与设备。
- **终端**：二进制 Protobuf WebSocket + node-pty + tmux，不解析 ANSI 来推断 Agent 语义。

同机 Caddy 的 443 使用设备 mTLS 与 bearer，8443 仅路由一次性配对；两者只绑定 Manager
主机已有 LAN 地址和 OpenVPN `10.66.0.1`。路由器仅可为 OpenVPN 转发 UDP 1194，不能转发
Manager/Console 的 TCP 端口。网络门禁见
[`deploy/manager/README.md`](../../deploy/manager/README.md)。

## 仓库

```text
proto/realtime/me/manager/  控制面与终端唯一契约
services/manager/           TypeScript Linux 服务、CLI adapter、SQLite、SSE/WS
apps/mobile/                统一 Flutter 手机客户端
packages/ag-ui-dart/        固定的 AG-UI Dart 最小协议 fork
deploy/manager/             Caddy、systemd、OpenVPN 端点 DDNS 与部署说明
docs/manager/research/      公开实现与复用决策
```

## 固定基线

- Node.js 24.18.0、pnpm 11.10.0
- Codex CLI 0.144.5
- Claude Code CLI 2.1.195
- Flutter 3.44.6 / Dart 3.12.2
- Linux 服务端；Android 是当前 MVP 客户端目标

Claude 的结构化 Ask 依赖固定 CLI 版本的私有 stdio control wire。版本或订阅认证不符合预期时服务端会关闭该 runtime，而不是尝试不受支持的兼容路径。

## 开发检查

```bash
corepack enable
pnpm install --frozen-lockfile
make generate
pnpm check
pnpm build
(cd apps/mobile && flutter analyze && flutter build apk --debug)
```

`make generate` 是 Protobuf/Pigeon 生成的唯一入口；不要手改 `services/manager/src/gen`、
`packages/manager-contracts-dart/lib/gen` 或生成的 Pigeon bridge。
升级固定 Codex CLI 时，使用
`pnpm --filter @realtime-me/manager generate:codex` 从该精确版本重新生成并规范化 app-server
契约；不要手改 adapter 的 `gen`/`schema` 目录。

## 本地服务

开发环境仍需要 OpenSSL；没有安装/登录 provider CLI 时服务可以启动，但相应 runtime 会显示不可用。

```bash
export SM_ALLOWED_WORKSPACE_ROOTS="$HOME/workspace"
export SM_SERVICE_URL="https://manager.realtime.internal"
export SM_CODEX_PATH=/absolute/path/to/codex
export SM_CLAUDE_PATH=/absolute/path/to/claude
pnpm --filter @realtime-me/manager dev
```

生产部署、PKI、Caddy、私有 DNS 和 OpenVPN 步骤见
[`deploy/manager/README.md`](../../deploy/manager/README.md)，整合约束见
[`project-consolidation.md`](../architecture/project-consolidation.md)。
