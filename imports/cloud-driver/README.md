# Cloud Drive

个人私有云盘，文件保存在部署主机本地 LVM 数据卷，通过 Cloudflare Pages 与 Tunnel 提供访问。

## 架构

- `web/apps/private`：带单密码登录的私有 Pages 应用。
- `web/apps/share`：仅用于只读分享链接的公开 Pages 应用。
- `api`：Go API、PostgreSQL 元数据和本地文件存储。
- `proto`：`cloud.drive.v1` 的唯一业务契约来源。
- `ops`：Docker Compose、Tunnel、LVM、备份和手动发布脚本。

实际域名、Cloudflare 项目名、Tunnel 标识、部署主机、卷组和容量均为部署时配置，
不写入版本控制。仓库中的 `example.com` 仅为文档占位域名。

私有 API 使用 bcrypt 密码校验和 24 小时签名会话 Cookie；公开分享 API 仅暴露解析、目录浏览、预览和下载路由。分享令牌本身是凭证，默认有效期为 7 天，最长 30 天。

## 本地开发

前端、后端和部署变量都通过相应的 `.env.example` 配置。不要提交真实域名、
Cloudflare 资源标识、主机信息、Tunnel token、数据库密码、登录密码、会话密钥或备份密码。

```sh
pnpm install
make generate
make verify
```

完整的主机初始化、Tunnel、Pages、密码配置、备份和手动发布步骤见 `ops/README.md`。

## 验证

本项目不维护单元测试文件。变更以 Proto lint/生成检查、Go 构建与 vet、前端类型检查/构建、Shell 静态检查及 Compose 配置渲染作为验证门禁。
