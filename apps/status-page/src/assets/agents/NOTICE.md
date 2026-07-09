# Agent animation assets

Neither agent's clips are in this repository. Both are the vendors' artwork, not
covered by this repository's MIT `LICENSE`, so neither is distributed with it.
`AgentCard.tsx` discovers both at build time: without them the build succeeds and
each agent falls back to an original mascot.

## `clawd/` — not in this repository

Claw'd is Anthropic's mascot, and Anthropic retains all rights in the artwork and
in the Claw'd and Claude marks.

To run the page with them, put Anthropic's clips through
`scripts/operator/normalize-clawd-clips.py`, which writes this directory. Each
clip draws Clawd at its own scale, so rendering them as published makes him
change size as the rotation advances; the tool normalises every clip on his eye
and emits the `<name>.png` posters the reduced-motion path needs. The names the
page looks for, and each clip's loop length, are listed in
`CLAWD_CLIP_DURATIONS_MS` in `AgentCard.tsx`. A clip missing either its GIF or
its poster is dropped; drop them all and the Claude agent falls back to
`agent-orbit.svg`.

## `codex/` — not in this repository

The Codex pets are OpenAI's artwork. The Codex CLI is Apache-2.0, but the pets
are not in that repository either: it downloads one spritesheet per pet from
OpenAI's CDN at runtime. Apache-2.0 cannot license files it does not contain,
and its §6 withholds trademark rights, so the sprites remain OpenAI brand assets.

`scripts/operator/normalize-codex-pets.py` fetches the eight spritesheets and
writes this directory. It slices row 7, the "running" animation, which is what
the pet does while Codex works, and stands every pet on the same baseline at one
shared scale so a pet and Claw'd render at one size. Every pet shares one loop
length, which is `CODEX_PET_DURATION_MS` in `AgentCard.tsx`. Drop the directory
and the Codex agent falls back to `codex-orbit.svg` and its siblings.

## `agent-orbit.svg` and `codex-*.svg`

Original work, covered by this repository's MIT `LICENSE`. These are the
fallbacks a clean checkout renders, and the only agent artwork committed here.
