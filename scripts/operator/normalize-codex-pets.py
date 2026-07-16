#!/usr/bin/env python3
"""Re-encode OpenAI's Codex pets as looping clips for the status page.

The pets are not shipped with the Codex CLI: it downloads one spritesheet per
pet from OpenAI's CDN and slices it at runtime. They are OpenAI's artwork, not
covered by that repository's Apache-2.0 licence, so like Anthropic's Claw'd
clips they are never committed here — see
apps/status-page/src/assets/agents/NOTICE.md.

Each sheet is a static 8x9 grid of 192x208 frames. Row 7 is the "running"
animation, which is the one that means the agent is working; its six frames run
at 120ms apiece with a 220ms hold on the last, exactly as
codex-rs/tui/src/pets/model.rs defines it. The page holds each frame twice that
long, because it draws the pet far larger than the terminal cell Codex sizes it
for, and the published rate reads as a blur at that size.

The pets are drawn at their true relative sizes and stand at different heights
inside the frame, so every clip is scaled by one shared factor rather than being
stretched to fill, and each pet is then dropped onto a common baseline. The
canvas matches the Claw'd clips, so a Codex pet and Claw'd render at one size.

The sprites carry no partial alpha, so a GIF's one transparent index reproduces
their edges exactly and the page needs no second image format. Their shading
does not survive a lossy encode: WebP rings around the hard pixel edges.
"""
from __future__ import annotations

import argparse
import sys
import urllib.request
from pathlib import Path

from PIL import Image

CDN_BASE_URL = "https://persistent.oaistatic.com/codex/pets/v1"
DOWNLOAD_USER_AGENT = "realtime-me-codex-pet-builder/1.0"
PETS = ["codex", "dewey", "fireball", "rocky", "seedy", "stacky", "bsod", "null-signal"]

FRAME_WIDTH = 192
FRAME_HEIGHT = 208
FRAME_COLUMNS = 8

# codex-rs/tui/src/pets/model.rs: app_state_animation(row_index=7, frame_count=6,
# frame_duration_ms=120, final_frame_duration_ms=220)
RUNNING_ROW = 7
RUNNING_FRAMES = 6
FRAME_DURATION_MS = 120
FINAL_FRAME_DURATION_MS = 220

# Codex runs the cycle inside a terminal cell. The page draws it many times that
# size, where a 120ms stride reads as a blur, so every frame is held twice as
# long. That puts a pet's stride at the cadence of Claw'd's own walk cycle.
FRAME_DURATION_SCALE = 2

# Shared with the Claw'd clips so both agents render at one size on one baseline.
CANVAS_WIDTH = 270
CANVAS_HEIGHT = 180
BASELINE_MARGIN = 2
PALETTE_COLORS = 128
TRANSPARENT_INDEX = 255


def main() -> int:
    parser = argparse.ArgumentParser(description="Write the Codex pet clips the status page discovers at build time.")
    parser.add_argument("target", type=Path, help="directory to write <pet>.gif and <pet>.png into")
    parser.add_argument("--spritesheets", type=Path, help="directory holding <pet>.webp sheets; downloaded when absent")
    args = parser.parse_args()

    # The sheets are OpenAI's, so they are cached outside the working tree.
    source = args.spritesheets or Path.home() / ".cache/realtime-me/codex-pets"
    source.mkdir(parents=True, exist_ok=True)
    args.target.mkdir(parents=True, exist_ok=True)

    durations = [FRAME_DURATION_MS * FRAME_DURATION_SCALE] * (RUNNING_FRAMES - 1) + [FINAL_FRAME_DURATION_MS * FRAME_DURATION_SCALE]
    scale = (CANVAS_HEIGHT - 2 * BASELINE_MARGIN) / FRAME_HEIGHT
    print(f"  canvas {CANVAS_WIDTH}x{CANVAS_HEIGHT}   scale {scale:.4f}   loop {sum(durations)}ms")
    print(f"  {'pet':14} {'content':>10} {'bytes':>8}")

    for pet in PETS:
        sheet = load_spritesheet(pet, source)
        frames = [crop_frame(sheet, RUNNING_ROW, column) for column in range(RUNNING_FRAMES)]
        box = union_box(frames)
        clips = [place(frame, box, scale) for frame in frames]

        clip_path = args.target / f"{pet}.gif"
        flattened = [quantize(clip) for clip in clips]
        flattened[0].save(
            clip_path,
            save_all=True,
            append_images=flattened[1:],
            duration=durations,
            loop=0,
            disposal=2,
            transparency=TRANSPARENT_INDEX,
            optimize=False,
        )
        poster(clips).save(args.target / f"{pet}.png", optimize=True)
        width = round((box[2] - box[0]) * scale)
        height = round((box[3] - box[1]) * scale)
        print(f"  {pet:14} {f'{width}x{height}':>10} {clip_path.stat().st_size // 1024:>7}K")

    print(f"\n  every clip loops in {sum(durations)}ms; CODEX_PET_DURATION_MS in AgentCard.tsx must match")
    return 0


def load_spritesheet(pet: str, source: Path) -> Image.Image:
    path = source / f"{pet}.webp"
    if not path.exists():
        url = f"{CDN_BASE_URL}/{pet}-spritesheet-v4.webp"
        request = urllib.request.Request(url, headers={"User-Agent": DOWNLOAD_USER_AGENT})
        with urllib.request.urlopen(request, timeout=60) as response:
            path.write_bytes(response.read())
    sheet = Image.open(path).convert("RGBA")
    expected = (FRAME_WIDTH * FRAME_COLUMNS, FRAME_HEIGHT * 9)
    if sheet.size != expected:
        raise SystemExit(f"{pet}: spritesheet is {sheet.size}, expected {expected}")
    return sheet


def crop_frame(sheet: Image.Image, row: int, column: int) -> Image.Image:
    left, top = column * FRAME_WIDTH, row * FRAME_HEIGHT
    return sheet.crop((left, top, left + FRAME_WIDTH, top + FRAME_HEIGHT))


def union_box(frames: list[Image.Image]) -> tuple[int, int, int, int]:
    box = None
    for frame in frames:
        found = frame.split()[-1].getbbox()
        if not found:
            continue
        box = found if box is None else (min(box[0], found[0]), min(box[1], found[1]), max(box[2], found[2]), max(box[3], found[3]))
    if box is None:
        raise SystemExit("a pet's running animation is entirely transparent")
    return box


def place(frame: Image.Image, box: tuple[int, int, int, int], scale: float) -> Image.Image:
    """Scale by the shared factor, centre the pet, and stand it on the baseline."""
    scaled = frame.resize((round(FRAME_WIDTH * scale), round(FRAME_HEIGHT * scale)), Image.LANCZOS)
    canvas = Image.new("RGBA", (CANVAS_WIDTH, CANVAS_HEIGHT), (0, 0, 0, 0))
    offset_x = round((CANVAS_WIDTH - (box[0] + box[2]) * scale) / 2)
    offset_y = round(CANVAS_HEIGHT - BASELINE_MARGIN - box[3] * scale)
    canvas.paste(scaled, (offset_x, offset_y), scaled)
    return canvas


def quantize(clip: Image.Image) -> Image.Image:
    """Flatten to a palette, reserving one index for the sprite's binary alpha."""
    alpha = clip.split()[-1]
    flat = clip.convert("RGB").quantize(colors=PALETTE_COLORS, method=Image.Quantize.MEDIANCUT)
    flat.paste(TRANSPARENT_INDEX, Image.eval(alpha, lambda value: 255 if value < 128 else 0))
    return flat


def poster(clips: list[Image.Image]) -> Image.Image:
    """The most opaque frame, so the reduced-motion still is never a blank one."""
    return max(clips, key=lambda clip: sum(clip.split()[-1].tobytes()))


if __name__ == "__main__":
    sys.exit(main())
