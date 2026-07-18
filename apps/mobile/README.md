# Super Manager Flutter client

当前 MVP 的 Android 客户端。它只连接 `smctl pair create` 生成的私有 CA/DDNS origin，使用设备 PKCS#12 与 bearer 完成双层认证。

```bash
flutter pub get
flutter analyze
flutter build apk --debug
```

主要页面：

- QR/粘贴配对与 Android secure storage；
- workspace、thread、runtime、quota 和设备管理；
- AG-UI timeline、结构化 Ask、cancel/steer 与 sequence replay；
- xterm.dart 原始 tmux 终端。

客户端不接受明文 HTTP，不包含跳过证书校验的开发开关，也不直接持有 OpenAI/Anthropic API key。
