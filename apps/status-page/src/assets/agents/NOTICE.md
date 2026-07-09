# Agent animation assets

## `clawd/` — not in this repository

Claw'd is Anthropic's mascot, and Anthropic retains all rights in the artwork and
in the Claw'd and Claude marks. The clips are **not** covered by this
repository's MIT `LICENSE`, so they are not distributed with it.

`AgentCard.tsx` discovers them at build time. Without them the build succeeds and
the Claude card falls back to `agent-orbit.svg`. To run the page with them, place
Anthropic's clips here as `clawd/<name>.gif`, each with a still frame beside it as
`clawd/<name>.png` for viewers who ask for reduced motion. The names and loop
lengths the card looks for are listed in `CLAWD_CLIPS` in `AgentCard.tsx`.

## `agent-orbit.svg` and `codex-*.svg`

Original work, covered by this repository's MIT `LICENSE`.
