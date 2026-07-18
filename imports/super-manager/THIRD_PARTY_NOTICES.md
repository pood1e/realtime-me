# Third-Party Notices

## AG-UI Dart SDK

- Project: AG-UI Protocol
- Source: <https://github.com/ag-ui-protocol/ag-ui/tree/main/sdks/community/dart>
- Imported release: `ag_ui 0.3.0`
- License: MIT; upstream text retained in [`packages/ag_ui/LICENSE`](packages/ag_ui/LICENSE)
- Local package version: `0.3.0+supermanager.1`
- Retained files: event/type models under `lib/src/events` and `lib/src/types`, bounded SSE parser under `lib/src/sse`, its two internal helpers, `lib/ag_ui.dart`, `LICENSE`, `CHANGELOG.md`, `analysis_options.yaml`, and `pubspec.yaml`
- Local modifications: added canonical Interrupt/resume/run outcome and capability fields needed by the app; removed the upstream HTTP client, encoder, backoff, examples, tests and their dependencies; narrowed exports to protocol models and SSE parsing

Super Manager supplies its own private-CA mTLS, bearer and sequence-replay transport. It does not retain or claim the upstream generic HTTP client.

## OpenAI Codex app-server protocol

- Project: OpenAI Codex
- Source: <https://github.com/openai/codex>
- Generator/runtime release: `@openai/codex 0.144.5`
- License: Apache-2.0; text retained in [`server/src/adapters/codex/LICENSE`](server/src/adapters/codex/LICENSE)
- Generated files: fixed app-server TypeScript declarations under `server/src/adapters/codex/gen` and JSON schemas under `server/src/adapters/codex/schema`
- Local processing: JSON object keys are sorted for reproducible output without changing schema fields; handwritten integration code stays outside the generated directories

Research-only references are recorded in [`docs/research/public-implementations.md`](docs/research/public-implementations.md). A reference is not a dependency and does not authorize copying its source.
