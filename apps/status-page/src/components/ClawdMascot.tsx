import type { ReactElement } from 'react';

// An original Clawd-style pixel mascot: a stout orange creature built entirely
// from <rect>s (the same rects-only approach the official mascot uses), animated
// with CSS. Not a copy of any Anthropic asset — a clean-room homage.
const BODY = '#d2703f';
const BODY_SHADE = '#b85a30';
const EYE = '#1c1917';

// o = body, d = shaded body, E = eye, . = transparent
const PIXELS = [
  '................',
  '................',
  '.....oooooo.....',
  '...oooooooooo...',
  '..oooooooooooo..',
  '.oooooooooooooo.',
  '.oooEEooooEEooo.',
  '.oooEEooooEEooo.',
  '.oooooooooooooo.',
  '.oooooooooooooo.',
  '.oooooooooooooo.',
  '..dddddddddddd..',
  '...dddddddddd...',
  '...dd.dddd.dd...',
  '................',
  '................',
];

export function ClawdMascot({ className = '' }: { className?: string }) {
  const rects: ReactElement[] = [];
  PIXELS.forEach((row, y) => {
    [...row].forEach((ch, x) => {
      if (ch === '.') return;
      const fill = ch === 'E' ? EYE : ch === 'd' ? BODY_SHADE : BODY;
      rects.push(
        <rect
          key={`${x}-${y}`}
          x={x}
          y={y}
          width={1.03}
          height={1.03}
          fill={fill}
          className={ch === 'E' ? 'clawd-eye' : undefined}
        />,
      );
    });
  });
  return (
    <svg viewBox="0 0 16 16" className={`clawd-svg ${className}`} shapeRendering="crispEdges" role="img" aria-label="Clawd working">
      <g className="clawd-body">{rects}</g>
    </svg>
  );
}
