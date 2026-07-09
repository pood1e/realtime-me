# Agent animation assets

## `clawd/` — not in this repository

Claw'd is Anthropic's mascot, and Anthropic retains all rights in the artwork and
in the Claw'd and Claude marks. The clips are **not** covered by this
repository's MIT `LICENSE`, so they are not distributed with it.

`AgentCard.tsx` discovers them at build time. Without them the build succeeds and
the Claude card falls back to `agent-orbit.svg`.

To run the page with them, put Anthropic's clips through
`scripts/operator/normalize-clawd-clips.py`, which writes this directory. Each
clip draws Clawd at its own scale, so rendering them as published makes him
change size as the card rotates; the tool normalises every clip on his eye and
emits the `<name>.png` posters the reduced-motion path needs. The names and loop
lengths the card looks for are listed in `CLAWD_CLIPS` in `AgentCard.tsx`.

## `agent-orbit.svg` and `codex-*.svg`

Original work, covered by this repository's MIT `LICENSE`.
