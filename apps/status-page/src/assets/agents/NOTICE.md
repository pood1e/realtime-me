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
emits the `<name>.png` posters the reduced-motion path needs. The names the page
looks for are listed in `CLAWD_TIER_CLIPS` in `AgentCard.tsx`, grouped by how
busy the agent is, and their loop lengths in `CLAWD_CLIP_DURATIONS_MS`. Every
tier must resolve, or the page falls back to `agent-orbit.svg` for all of them.

## `agent-orbit.svg` and `codex-*.svg`

Original work, covered by this repository's MIT `LICENSE`. OpenAI publishes no
mascot to normalise the way Anthropic's clips are, so Codex's four busyness
tiers are drawn here as animated SVG rather than shipped as artwork.
