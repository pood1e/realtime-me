# Realtime Me Manager AG-UI Dart protocol package

这是从 AG-UI community Dart SDK 0.3.0 固定并裁剪出的内部包，不发布到 pub.dev。

保留内容：

- canonical AG-UI event、message、tool、context 和 capability 模型；
- Interrupt、resume 与 run-finished outcome；
- 有界 SSE message parser。

上游通用 HTTP client、重试和状态管理已删除。Realtime Me Manager 必须使用自己的私有
CA/mTLS、bearer、sequence replay transport；项目中不得再引入第二套 AG-UI wire model。

许可与本地修改见根目录 [`THIRD_PARTY_NOTICES.md`](../../THIRD_PARTY_NOTICES.md)。
