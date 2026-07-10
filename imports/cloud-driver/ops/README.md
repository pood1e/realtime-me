# 生产运维

本目录只部署主机侧的 `api`、PostgreSQL 与 `cloudflared`。前端是两个独立的 Cloudflare Pages 静态项目；它们不在 Docker Compose 中运行。所有生产命令均在 `<DEPLOY_USER>@<DEPLOY_HOST>` 上以 root 执行，推荐将仓库检出到 `/opt/cloud-drive`。

## 容器边界

- `postgres` 仅加入内部 `backend` 网络，未发布主机端口。
- `api` 仅通过 `edge` Docker 网络被 `cloudflared` 访问，未发布主机端口。
- `cloudflared` 建立出站 Tunnel；部署包装脚本仅短暂读取 root-only token 文件，并以 Docker secret 挂载给 UID/GID `65532` 的容器，不出现在命令行、`docker inspect` 或 Compose 环境文件中。
- Compose 使用 `bind.create_host_path: false`。未挂载数据卷或目录缺失时部署必须失败，而不是在根文件系统悄然创建目录。

## 从 staging 安装

在部署主机上，将已准备好的 staging 工作树安装到 `/opt/cloud-drive`，并将 Tunnel token 直接复制为 root-only 文件；命令不会打印 token 内容：

```bash
sudo install -d -o root -g root -m 0755 /opt/cloud-drive
sudo rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w /path/to/cloud-drive-staging/ /opt/cloud-drive/
sudo install -d -o root -g root -m 0700 /etc/cloud-drive
sudo install -o root -g root -m 0400 \
  /path/to/cloud-drive-secrets/cloudflare-tunnel.token \
  /etc/cloud-drive/cloudflare-tunnel.token
```

## 首次主机初始化

先安装 Docker Engine（含 Compose v2）、LVM2、`rsync`、`util-linux` 与 `openssl`，并确认目标卷组有足够可用空间。以下脚本会创建并格式化一个**新的**逻辑卷；它拒绝复用已有卷。卷组名和容量必须显式提供，不保存在仓库中。

```bash
sudo /opt/cloud-drive/ops/scripts/initialize-storage.sh \
  --vg-name <VOLUME_GROUP> \
  --size <LOGICAL_VOLUME_SIZE>
```

脚本将创建 `/srv/cloud-drive/data`，以 `nodev,nosuid,noexec,noatime` 写入 `/etc/fstab`，并准备：

- `files/`：API 容器（UID/GID `10001`）的可写数据根目录；API 在其中管理 `blobs/` 和临时 `uploads/`。
- `postgres/`：`postgres:18.4-alpine`（UID/GID `70`）的数据目录。
- `backup-staging/`：仅 root 可读写的短期 `pg_dump` 暂存目录。

USB 备份盘必须是已有的 ext4 文件系统。先人工核对设备与 UUID，然后只添加 UUID 挂载（脚本绝不格式化 USB 盘）：

```bash
sudo /opt/cloud-drive/ops/scripts/configure-backup-disk.sh --uuid <USB_FILESYSTEM_UUID>
```

它会把磁盘挂载到 `/mnt/cloud-drive-backup`，将文件系统根目录收紧为 `root:root`/`0711`，并以 `nodev,nosuid,noexec,nofail` 写入 `/etc/fstab`。已有子目录的所有权和内容不会被修改。`nofail` 只允许主机正常启动；备份脚本会把 USB 未挂载或挂载根目录可被非 root 替换视为明确失败。

## 私密运行时配置

运行时配置从不提交到仓库。创建 `/etc/cloud-drive/runtime.env`：

```bash
sudo install -d -o root -g root -m 0700 /etc/cloud-drive
sudo install -o root -g root -m 0600 /opt/cloud-drive/ops/.env.example /etc/cloud-drive/runtime.env
sudoedit /etc/cloud-drive/runtime.env
```

将示例域名替换为同一 Cloudflare zone 中的四个实际域名：

| 用途 | 示例值 |
| --- | --- |
| 私有 Pages | `https://drive.example.com` |
| 私有 API | `drive-api.example.com` |
| 公开分享 Pages | `https://share.example.com` |
| 公开分享 API | `share-api.example.com` |

将 `POSTGRES_PASSWORD` 替换为 `openssl rand -hex 32` 的输出。该格式可以安全地嵌入 Compose 构造的 `postgres://` URL。

在可信开发机上交互式输入云盘密码并生成 bcrypt cost 12 哈希；密码不会进入命令行参数。把输出写入 `PASSWORD_HASH_BASE64`，再将 `openssl rand -hex 32` 的输出写入 `SESSION_SECRET`：

```bash
htpasswd -nBC 12 cloud-drive | cut -d: -f2 | base64 | tr -d '\n'
openssl rand -hex 32
```

API 只读取 bcrypt 哈希，不保存明文密码。成功登录后签发 24 小时、host-only、`HttpOnly`、`Secure`、`SameSite=Strict` 的签名 Cookie。登录在执行 bcrypt 前按 Cloudflare 客户端 IP 限制为滚动 60 秒最多 5 次，并将 bcrypt 并发硬限制为 2，状态表最多保留 1024 个客户端。

Tunnel token 必须单独保存为仅一行、root-only 文件，不能写入 `runtime.env`：

```bash
sudo install -o root -g root -m 0400 \
  /path/to/cloud-drive-secrets/cloudflare-tunnel.token \
  /etc/cloud-drive/cloudflare-tunnel.token
```

`CLOUDFLARE_TUNNEL_TOKEN_FILE` 保持指向该文件。`deploy.sh`、`backup.sh` 和 `compose.sh` 仅在其自身进程环境中短暂将 token 交给 Compose；Compose 以仅授予 `cloudflared` 的 `/run/secrets/cloudflare_tunnel_token`（UID/GID `65532`、模式 `0400`）挂载它。需要 `cloudflared` 2025.4.0 或更新版本。

## Cloudflare Tunnel 与 Pages

先按自己的 zone、Pages 项目名和 Tunnel ID 创建以下四条 Cloudflare DNS CNAME 记录。表中值均为占位示例：

| 名称 | 类型 | 目标 | 代理状态 |
| --- | --- | --- | --- |
| `drive.example.com` | CNAME | `<PRIVATE_PAGES_PROJECT>.pages.dev` | 已代理（橙云） |
| `share.example.com` | CNAME | `<SHARE_PAGES_PROJECT>.pages.dev` | 已代理（橙云） |
| `drive-api.example.com` | CNAME | `<TUNNEL_ID>.cfargotunnel.com` | 已代理（橙云） |
| `share-api.example.com` | CNAME | `<TUNNEL_ID>.cfargotunnel.com` | 已代理（橙云） |

若同名的 A、AAAA 或旧 CNAME 记录已存在，先删除或替换它；CNAME 不能与这些同名记录共存。

前两条 CNAME 生效后，已添加到 Pages 项目的自定义域会从 `pending` 转为 `active`；不要重复添加域名。

1. 在 Cloudflare Dashboard 创建一个 **remotely managed named tunnel**，取得其 token 并仅写入上述 root-only token 文件。
2. 在该 Tunnel 的 Public Hostnames 中配置：
   - `drive-api.example.com` → `http://api:8080`
   - `share-api.example.com` → `http://api:8080`
   - 添加最终 catch-all `http_status:404`，避免未知主机名被转发。
   `api` 是同一 Compose `edge` 网络中的服务名，不是主机端口。
3. 在 Cloudflare Pages 创建两个静态项目，绑定 `drive.example.com` 与 `share.example.com`。私有前端只请求 `https://drive-api.example.com`，公开分享前端只请求 `https://share-api.example.com`。
4. 不要在 `drive.example.com` 或 `drive-api.example.com` 前添加 Cloudflare Access。私有 Pages 只包含静态登录界面，不含用户数据；`driver-api` 只公开健康检查与登录 RPC，其余私有 ConnectRPC、上传和下载路由全部要求有效会话 Cookie。API 严格校验 `Origin`，只允许 `https://drive.example.com` 进行带凭据跨域请求。
5. 为 `share-api.example.com` 设置 Cache Rule 为 bypass/no-store，并配置合理的 WAF/速率限制。API 按 Host 区分私有与分享路由，分享端只能读取令牌授权范围。

Tunnel 为出站连接，主机不应为 API、PostgreSQL 或 `cloudflared` 新增 HTTP 入站防火墙规则。Cloudflare 的远程 Tunnel token 与 token-file 参数说明见 [Cloudflare Tunnel run parameters](https://developers.cloudflare.com/tunnel/advanced/run-parameters/)。

## 手动发布与运行检查

每次后端发布前先更新 `/opt/cloud-drive` 工作树，再执行：

```bash
sudo /opt/cloud-drive/ops/scripts/deploy.sh
```

该脚本会校验 root-only 配置、数据卷标记、20 GiB 本地安全余量、Compose 配置和服务健康状态，然后运行 `docker compose up --build --detach --wait`。`cloudflared` 只有在其 `/ready` 指标端点确认至少一条 Tunnel 连接后才会变为 healthy；它不会显示运行时秘密。排障时可使用：

```bash
sudo /opt/cloud-drive/ops/scripts/compose.sh -- ps
sudo /opt/cloud-drive/ops/scripts/compose.sh -- logs --tail=100 cloudflared
```

在已通过 Wrangler 交互式登录的开发机上手动部署两个 Pages 产物：

```bash
/path/to/cloud-drive/ops/scripts/deploy-pages.sh
```

先复制 `ops/pages.env.example` 为被忽略的 `ops/pages.env`，填入 Pages 项目名和 API 地址。脚本会构建 `web/apps/private/dist` 与 `web/apps/share/dist`，生成与 API Origin 匹配的 CSP，再发布到配置的两个 Pages 项目。首次使用时先在 Cloudflare 创建项目并绑定上文域名。不要把真实域名、项目名或 Cloudflare API token 提交到仓库。

## 持久化运维权限

不要把部署用户加入 `docker` 组；Docker socket 等价于主机 root 权限。首次安装完成后，可由 root 为现有 SSH 用户安装受限运维网关：

```bash
sudo /opt/cloud-drive/ops/scripts/install-operator-access.sh --user <DEPLOY_USER>
```

脚本创建 `cloud-drive-operators` 组，但登录用户仍是普通非 root 用户。重新登录后，该组只能免密码调用以下四个 root-owned 固定入口，不能运行任意 sudo 或直接访问 Docker socket：

```bash
sudo -n cloud-drive-release-api
sudo -n cloud-drive-backup-now
sudo -n cloud-drive-status
sudo -n cloud-drive-logs api        # api | postgres | cloudflared | backup
```

常规 API 发布先从 Git 提交生成干净的 staging，再触发发布入口：

```bash
stage=$(mktemp -d)
git archive HEAD api | tar -x -C "$stage"
rsync -rlt --delete --omit-dir-times --exclude=/Dockerfile \
  "$stage/api/" <DEPLOY_USER>@<DEPLOY_HOST>:/var/lib/cloud-drive-release/incoming-api/
ssh <DEPLOY_USER>@<DEPLOY_HOST> sudo -n cloud-drive-release-api
rm -rf "$stage"
```

发布入口只更新 `api/` 源码，固定使用 root-owned Dockerfile、Compose、部署脚本和运行时配置；构建失败时自动恢复上一份 API 源码并重新部署。前端仍由开发机直接发布 Pages，不需要主机权限。Dockerfile、Compose、systemd 或 `ops/` 变更属于低频控制面操作，仍需显式 root 审核并重新安装相关入口。

## 明文增量备份

创建 root-only 的备份配置：

```bash
sudo install -o root -g root -m 0600 /dev/null /etc/cloud-drive/backup.env
sudoedit /etc/cloud-drive/backup.env
```

`backup.env` 使用如下无引号的键值；`BACKUP_ROOT_DIR` 必须位于 USB 挂载点内：

```dotenv
BACKUP_MOUNT_DIR=/mnt/cloud-drive-backup
BACKUP_ROOT_DIR=/mnt/cloud-drive-backup/cloud-drive-backups
```

启用每日定时任务，并立即执行一次以验证完整备份链路：

```bash
sudo /opt/cloud-drive/ops/scripts/install-backup-timer.sh
sudo systemctl start cloud-drive-backup.service
sudo systemctl status --no-pager cloud-drive-backup.service
sudo systemctl list-timers cloud-drive-backup.timer
sudo readlink -f /mnt/cloud-drive-backup/cloud-drive-backups/latest
```

定时器每天运行一次，并在六小时窗口内随机延迟。每个成功快照位于 `cloud-drive-backups/snapshots/<UTC时间>/`，包含可直接读取的 `blobs/` 与未加密的 PostgreSQL custom dump `postgres.dump`；`latest` 原子指向最新成功快照。未变化的 blob 通过 ext4 硬链接复用空间，保留最近 30 个成功快照。临时 `files/uploads/` 分片不会备份。

USB 未挂载、根配置不安全、源/备份卷低于 20 GiB 安全余量、首次备份容量明显不足或任一 `rsync`/数据库命令失败都会让 systemd 单元失败并写入 journal；未完成快照会清理，不会替换 `latest`。这些快照没有加密，能够读取移动硬盘的人也能读取全部文件和数据库备份。

恢复演练应先从 `latest` 复制到隔离目录，使用 `pg_restore --list postgres.dump` 验证数据库归档，再测试一个 blob 与对应元数据。正式恢复时先停止写入，在维护窗口中恢复 PostgreSQL dump 与同一快照中的 `blobs/`，不要混用不同日期的内容。
