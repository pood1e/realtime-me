# 生产运维

本目录只管理主机上的 PostgreSQL、一次性迁移、Go API 与内容 Worker。共享 Tunnel
connector 由 [`deploy/edge`](../edge/README.md) 管理，七个 Cloudflare Pages 项目由
[`deploy/web`](../web/README.md) 发布。所有真实域名、主机地址和凭据都留在被 Git 忽略的
本地配置或主机 root-only 文件中。

## 容器边界

- `postgres` 只加入内部 `backend` 网络，不发布端口。
- `migrate` 只加入 `backend`，在每次发布时先校验并应用 append-only 迁移，成功退出后
  API 与 Worker 才能启动。
- `worker` 加入内部 `backend` 与独立 `provider-egress` 网络；前者访问 PostgreSQL，
  后者只用于歌单音频出站下载。Worker 不监听 HTTP 端口。
- `api` 加入 `backend` 与外部 `realtime-me-edge`，稳定别名为 `library-api`；共享
  `cloudflared` connector 只通过该别名访问 API。
- 所有 bind mount 都使用 `create_host_path: false`；数据卷未挂载时部署直接失败。
- API、Worker 与 PostgreSQL 使用 `restart: unless-stopped`；Migrate 是
  `restart: no` 的启动门禁。

## 安装工作树

推荐将审核后的工作树安装到 `/opt/cloud-drive`：

```bash
sudo install -d -o root -g root -m 0755 /opt/cloud-drive
sudo rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w \
  /path/to/cloud-drive-staging/ /opt/cloud-drive/
sudo install -d -o root -g root -m 0700 /etc/cloud-drive
```

首次部署前先启动 `deploy/edge`，由它创建外部 `realtime-me-edge` network。Library 的
运行时环境不读取 Tunnel token，也不能管理 edge connector。

## 初始化主数据卷

主机需要 Docker Engine + Compose v2、LVM2、rsync、util-linux、openssl 与
apache2-utils。初始化脚本只创建新的 ext4 逻辑卷，拒绝格式化已存在的卷：

```bash
sudo /opt/cloud-drive/deploy/library/scripts/initialize-storage.sh \
  --vg-name <VOLUME_GROUP> \
  --size <LOGICAL_VOLUME_SIZE>
```

默认挂载点为 `/srv/cloud-drive/data`，并通过 UUID 写入 `/etc/fstab`。目录所有权：

- `files/`：UID/GID `10001`，存放 `objects/`、`artifacts/`、`uploads/` 与 `work/`；
- `postgres/`：PostgreSQL Alpine UID/GID `70`；
- `backup-staging/`：root-only 的临时数据库 dump。

USB 备份盘必须已经是独立 ext4 文件系统。人工确认 UUID 后执行：

```bash
sudo /opt/cloud-drive/deploy/library/scripts/configure-backup-disk.sh \
  --uuid <USB_FILESYSTEM_UUID>
```

该脚本不会格式化 USB，也不会修改已有子目录内容。`nofail` 只保证主机能够启动；
备份时如果 USB 没有真正挂载，任务一定失败。

## 运行时配置

```bash
sudo install -o root -g root -m 0600 \
  /opt/cloud-drive/deploy/library/.env.example \
  /etc/cloud-drive/runtime.env
sudoedit /etc/cloud-drive/runtime.env
```

关键变量：

| 变量 | 内容 |
| --- | --- |
| `PRIVATE_APP_ORIGINS` | 认证、云盘、书架、音乐盒、图床的逗号分隔 HTTPS Origin |
| `PUBLIC_APP_ORIGINS` | 壁纸站与分享页的逗号分隔 HTTPS Origin |
| `SHARE_APP_ORIGIN` | 创建文件分享链接时使用的分享页 Origin |
| `MUSIC_APP_ORIGIN` | Spotify OAuth 完成后固定返回的音乐盒 Origin |
| `PRIVATE_API_HOST` | 所有私有应用共用的 API Host，仅主机名 |
| `PUBLIC_API_HOST` | 壁纸、匿名图片与分享共用的 API Host，仅主机名 |
| `PASSWORD_HASH_BASE64` | bcrypt cost 12 以上的密码哈希再做 padded Base64 |
| `SESSION_SECRET` | 64 个十六进制字符 |
| `MUSIC_PROVIDER_CREDENTIAL_KEY` | 32 字节 padded Base64，仅加密第三方账号凭据 |
| `RESERVED_FREE_BYTES` | 上传和歌单下载必须共同保留的本地空闲字节数，默认 20 GiB |
| `SPOTIFY_CLIENT_ID` / `SPOTIFY_CLIENT_SECRET` | 可选，但必须同时配置 |

生成运行时秘密：

```bash
htpasswd -nBC 12 local-library | cut -d: -f2 | base64 | tr -d '\n'
openssl rand -hex 32
openssl rand -base64 32  # 音乐 Provider 凭据密钥
openssl rand -hex 32  # PostgreSQL 密码
```

API 不保存明文密码。登录会话有效 24 小时，Cookie 为 host-only、`HttpOnly`、
`Secure`、`SameSite=Strict`。前端只在认证 Pages 提交密码；其他私有 Pages 未登录时
跳转过去。

Tunnel token 单独保存：

```bash
sudo install -o root -g root -m 0400 \
  /path/to/cloudflare-tunnel.token \
  /etc/cloud-drive/cloudflare-tunnel.token
```

## Tunnel 与 DNS

在一条 remotely-managed named tunnel 中配置两个 Public Hostname：

```text
<PRIVATE_API_HOST> -> http://api:8080
<PUBLIC_API_HOST>  -> http://api:8080
catch-all          -> http_status:404
```

两个 API DNS 名均指向该 Tunnel；七个前端 DNS 名分别绑定七个 Pages 项目。不要给
私有 Pages 或私有 API 添加 Cloudflare Access：静态应用没有用户数据，真正的读取与
写入都由 API 的密码会话和精确 Origin 校验保护。

私有 API 只公开登录和健康 RPC；其余 ConnectRPC、上传、源文件与衍生文件路由均需
有效 Cookie。公开 API 只挂载壁纸读取、匿名图片和令牌范围内的只读分享路由。
Spotify OAuth callback 是私有 Host 上唯一额外的无会话 GET 路由，通过一次性 state
与 PKCE 验证，不依赖 `SameSite=Strict` Cookie，也不接受客户端指定的跳转 Origin。

## 启动与更新主机服务

```bash
sudo /opt/cloud-drive/deploy/library/scripts/deploy.sh
sudo /opt/cloud-drive/deploy/library/scripts/compose.sh -- ps
sudo /opt/cloud-drive/deploy/library/scripts/compose.sh -- logs --tail=100 api worker
```

`deploy.sh` 校验 root-only 配置、挂载标记、20 GiB 余量与 Compose 渲染，然后构建
同一个包含 Migrate、API 和 Worker 二进制的镜像并等待服务健康。Worker 通过数据库
心跳上报状态，并用独立的交互任务通道和 Provider 下载通道处理工作；重启后会凭租约
继续认领未完成任务。

## 发布七个 Pages

Web 发布已从主机运行时中拆出。使用 `deploy/web/deploy-library-pages.sh` 和
`deploy/web/pages.env`；完整命令见 [`deploy/web/README.md`](../web/README.md)。发布 Web
不会触发数据库迁移、备份或 Library 容器重启。

## 受限非 root 运维入口

不要把部署用户加入 `docker` 组。由 root 安装固定 sudo 网关：

```bash
sudo /opt/cloud-drive/deploy/library/scripts/install-operator-access.sh \
  --user <DEPLOY_USER>
```

重新登录后该用户只能执行：

```bash
sudo -n cloud-drive-release
sudo -n cloud-drive-backup-now
sudo -n cloud-drive-status
sudo -n cloud-drive-logs api       # api | worker | migrate | postgres | backup
```

`sudo -n` 只进入 root 安装的固定网关，不会询问密码，也不会授予 Docker socket 或
通用 sudo 权限。发布前同时暂存 API/Worker 源码与 Compose：

```bash
rsync -rltzO --delete --exclude=/Dockerfile \
  services/library/ /var/lib/cloud-drive-release/incoming-source/
cp deploy/library/compose.yaml \
  /var/lib/cloud-drive-release/incoming-compose/compose.yaml
sudo -n cloud-drive-release
```

唯一发布网关会把源码、迁移和 Compose 作为一个版本串行部署。网关先使用无敏感信息的
环境渲染 Compose，再按固定的服务、镜像、网络和挂载策略校验真实配置，并在
越过迁移边界前强制生成一份成功备份。当前策略只允许 PostgreSQL、一次性 Migrate、API、
本机下载 Worker；`edge` 必须是名为 `realtime-me-edge` 的外部网络。数据库迁移是
forward-only；如果迁移后的启动失败，候选
版本会保留以便向前修复，不会用不兼容的旧程序覆盖新 schema。Dockerfile、运维脚本或
Compose 安全策略本身仍属于 root 控制面，需要 root 重新审核安装。根 `go.mod`、`go.sum`、
`vendor/`、Library 生成的 Go contracts 与 Dockerfile 同样由 root 控制；依赖或 Proto 变更
必须先由 root 安装新的整合工作树，受限发布入口只替换 `services/library` 业务源码。

## 明文增量备份

创建 `/etc/cloud-drive/backup.env`：

```dotenv
BACKUP_MOUNT_DIR=/mnt/cloud-drive-backup
BACKUP_ROOT_DIR=/mnt/cloud-drive-backup/cloud-drive-backups
```

文件必须为 root:root、`0600`。启用定时器并立即验证：

```bash
sudo /opt/cloud-drive/deploy/library/scripts/install-backup-timer.sh
sudo systemctl start cloud-drive-backup.service
sudo systemctl status --no-pager cloud-drive-backup.service
sudo readlink -f /mnt/cloud-drive-backup/cloud-drive-backups/latest
```

每份成功快照包含：

```text
snapshots/<UTC>/
├── manifest
├── postgres.dump
└── objects/
```

清单记录 dump SHA-256、对象数量和逻辑字节数，创建后还会以 checksum dry-run 比对源
对象。未变化对象通过 ext4 hard link 与上一份快照复用，默认保留最近 30 份。`artifacts/`、
`uploads/`、`work/` 不备份；前者可重建，后两者是临时状态。备份不加密，能够读取
移动硬盘的人可以直接读取内容与数据库 dump；第三方音乐账号字段仍是应用层密文。
`MUSIC_PROVIDER_CREDENTIAL_KEY` 必须在数据库之外单独保管，否则恢复后需要重新连接
QQ 音乐、网易云和 Spotify。

## 恢复演练

先验证归档可读：

```bash
pg_restore --list /path/to/snapshot/postgres.dump >/dev/null
```

正式恢复会替换当前数据库和源文件，必须在维护窗口显式确认：

```bash
sudo /opt/cloud-drive/deploy/library/scripts/restore.sh \
  --snapshot /mnt/cloud-drive-backup/cloud-drive-backups/latest \
  --confirm-destroy
```

恢复脚本停止写入服务、恢复同一快照的数据库和 `objects/`、删除衍生缓存、为全部内容
重建处理队列，再启动完整 Compose。不要混用不同日期的数据库和对象目录。
