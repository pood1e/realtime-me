#!/usr/bin/env python3
"""Re-encode Anthropic's Claw'd clips so the mascot renders at one consistent size.

The clips are not in this repository (see
apps/web/status/src/assets/agents/NOTICE.md). Each one draws Clawd at a
different scale — his eye is 29px across in Laptop and 100px in CrabWalking —
inside a differently-padded canvas, so `object-fit: contain` in the card's fixed
box showed him anywhere from tiny to full-bleed. The eye is a perfect square in
every clip and does not change shape with the pose, so it is the scale reference:
normalise every clip until the eye is the same size, then bottom-align each
clip's motion envelope on a shared baseline in one common canvas. Frame timings
and transparency are preserved; identical consecutive frames are coalesced.

Needs Pillow and numpy. Writes <name>.gif plus a <name>.png poster, the most
opaque frame, shown to viewers who ask for reduced motion.

  scripts/operator/normalize-clawd-clips.py \
    ~/Downloads/clawd-gifs apps/web/status/src/assets/agents/clawd
"""
from __future__ import annotations

import statistics
import sys
from pathlib import Path

import numpy as np
from PIL import Image, ImageSequence

BODY_RGB = np.array([217, 119, 87])
CANVAS_H = 180
BASELINE_MARGIN = 2
PALETTE_COLORS = 128
TRANSPARENT_INDEX = 255

CLIPS = {
    "Clawd-Laptop.gif": "clawd-laptop",
    "Clawd-Magnifier.gif": "clawd-magnifier",
    "Clawd-CrabWalking.gif": "clawd-crab-walking",
    "Clawd-Lurking.gif": "clawd-lurking",
    "Clawd-RacingCar.gif": "clawd-racing-car",
    "Clawd-Soccer.gif": "clawd-soccer",
    "Clawd-Dancing.gif": "clawd-dancing",
    "Clawd-JumpingHappy.gif": "clawd-jumping-happy",
    "Clawd-Waving.gif": "clawd-waving",
}


def frames(path: Path) -> list[tuple[Image.Image, int]]:
    image = Image.open(path)
    return [(frame.convert("RGBA"), frame.info.get("duration", 100)) for frame in ImageSequence.Iterator(image)]


def dark_components(mask: np.ndarray) -> list[tuple[int, int, int]]:
    height, width = mask.shape
    seen = np.zeros_like(mask, dtype=bool)
    found = []
    for y in range(height):
        for x in range(width):
            if not mask[y, x] or seen[y, x]:
                continue
            stack = [(y, x)]
            seen[y, x] = True
            ys, xs = [], []
            while stack:
                cy, cx = stack.pop()
                ys.append(cy)
                xs.append(cx)
                for dy, dx in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                    ny, nx = cy + dy, cx + dx
                    if 0 <= ny < height and 0 <= nx < width and mask[ny, nx] and not seen[ny, nx]:
                        seen[ny, nx] = True
                        stack.append((ny, nx))
            found.append((max(xs) - min(xs) + 1, max(ys) - min(ys) + 1, len(xs)))
    return found


def eye_size(clip: list[tuple[Image.Image, int]]) -> float:
    """Median width of Clawd's eye, the one art element every pose shares."""
    widths = []
    for frame, _ in clip:
        # Lurking opens on transparent frames, so scan until enough eyes are seen
        # rather than trusting a fixed window at the head of the clip.
        if len(widths) >= 20:
            break
        pixels = np.array(frame)
        rgb = pixels[..., :3].astype(int)
        alpha = pixels[..., 3]
        body = (np.abs(rgb - BODY_RGB).max(axis=2) < 45) & (alpha > 128)
        if body.sum() < 400:
            continue
        ys, xs = np.where(body)
        crop = pixels[ys.min() : ys.max() + 1, xs.min() : xs.max() + 1]
        step = max(1, crop.shape[1] // 240)
        crop = crop[::step, ::step]
        dark = (crop[..., :3].max(axis=2) < 70) & (crop[..., 3] > 128)
        if not dark.any():
            continue
        body_area = body.sum() / (step * step)
        eyes = [
            component
            for component in dark_components(dark)
            if component[2] < 0.03 * body_area and component[0] > 1 and component[1] > 1 and 0.5 < component[0] / component[1] < 2.0
        ]
        eyes.sort(key=lambda component: -component[2])
        widths.extend(component[0] * step for component in eyes[:2])
    if not widths:
        raise SystemExit("no eye found")
    return statistics.median(widths)


def union_box(clip: list[tuple[Image.Image, int]]) -> tuple[int, int, int, int]:
    box = None
    for frame, _ in clip:
        found = frame.split()[-1].getbbox()
        if not found:
            continue
        box = found if box is None else (min(box[0], found[0]), min(box[1], found[1]), max(box[2], found[2]), max(box[3], found[3]))
    return box


def quantize(frame: Image.Image) -> Image.Image:
    alpha = frame.split()[-1]
    flat = frame.convert("RGB").quantize(colors=PALETTE_COLORS, method=Image.Quantize.MEDIANCUT)
    flat.paste(TRANSPARENT_INDEX, Image.eval(alpha, lambda value: 255 if value < 128 else 0))
    return flat


def main() -> int:
    source = Path(sys.argv[1]).expanduser()
    target = Path(sys.argv[2]).expanduser()
    target.mkdir(parents=True, exist_ok=True)

    loaded = {name: frames(source / name) for name in CLIPS}
    eyes = {name: eye_size(clip) for name, clip in loaded.items()}
    boxes = {name: union_box(clip) for name, clip in loaded.items()}

    # Solve for the eye size that makes the tallest normalised clip exactly fill
    # the canvas height, so nothing is cropped and nothing is needlessly small.
    def height_at(eye_px: float, name: str) -> float:
        scale = eye_px / eyes[name]
        box = boxes[name]
        return (box[3] - box[1]) * scale

    eye_px = 100.0
    tallest = max(height_at(eye_px, name) for name in CLIPS)
    eye_px *= (CANVAS_H - 2 * BASELINE_MARGIN) / tallest

    widths = []
    for name in CLIPS:
        scale = eye_px / eyes[name]
        box = boxes[name]
        widths.append((box[2] - box[0]) * scale)
    canvas_w = int(max(widths)) + 2
    canvas_w += canvas_w % 2

    print(f"  normalised eye = {eye_px:.1f}px   canvas = {canvas_w}x{CANVAS_H}")
    print(f"  {'clip':22} {'source eye':>10} {'scale':>7} {'content':>12}")

    for name, stem in CLIPS.items():
        clip = loaded[name]
        scale = eye_px / eyes[name]
        box = boxes[name]
        content_w = max(1, round((box[2] - box[0]) * scale))
        content_h = max(1, round((box[3] - box[1]) * scale))
        offset_x = (canvas_w - content_w) // 2
        offset_y = CANVAS_H - BASELINE_MARGIN - content_h

        out_frames, durations = [], []
        for frame, duration in clip:
            cropped = frame.crop(box).resize((content_w, content_h), Image.LANCZOS)
            canvas = Image.new("RGBA", (canvas_w, CANVAS_H), (0, 0, 0, 0))
            canvas.paste(cropped, (offset_x, offset_y), cropped)
            out_frames.append(quantize(canvas))
            durations.append(duration)

        out = target / f"{stem}.gif"
        out_frames[0].save(
            out,
            save_all=True,
            append_images=out_frames[1:],
            duration=durations,
            loop=0,
            disposal=2,
            transparency=TRANSPARENT_INDEX,
            optimize=False,
        )
        print(f"  {stem:22} {eyes[name]:10.0f} {scale:7.3f} {f'{content_w}x{content_h}':>12}  {out.stat().st_size // 1024}K")

        # Poster: the most opaque frame, never a blank one.
        best, coverage = None, -1.0
        for frame, _ in clip:
            cropped = frame.crop(box).resize((content_w, content_h), Image.LANCZOS)
            canvas = Image.new("RGBA", (canvas_w, CANVAS_H), (0, 0, 0, 0))
            canvas.paste(cropped, (offset_x, offset_y), cropped)
            share = sum(canvas.split()[-1].tobytes()) / (canvas_w * CANVAS_H * 255)
            if share > coverage:
                coverage, best = share, canvas
        assert coverage > 0.03, f"{stem} poster is blank ({coverage:.3%})"
        best.save(target / f"{stem}.png", optimize=True)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
