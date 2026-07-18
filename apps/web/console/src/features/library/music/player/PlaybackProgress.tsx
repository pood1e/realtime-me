import { clock } from "../music-model";

export function PlaybackProgress({
  position,
  duration,
  onSeek,
}: {
  position: number;
  duration: number;
  onSeek: (seconds: number) => void;
}) {
  const boundedDuration = Number.isFinite(duration) ? Math.max(0, duration) : 0;
  const boundedPosition = Math.min(
    boundedDuration,
    Number.isFinite(position) ? Math.max(0, position) : 0,
  );
  return (
    <div className="flex items-center gap-2 sm:gap-3">
      <span className="w-9 text-right text-[10px] text-muted-foreground sm:w-10 sm:text-[11px]">
        {clock(boundedPosition)}
      </span>
      <input
        aria-label="播放进度"
        type="range"
        min={0}
        max={boundedDuration}
        step={0.1}
        value={boundedPosition}
        disabled={boundedDuration <= 0}
        onChange={(event) => onSeek(Number(event.target.value))}
        className="h-4 min-w-0 flex-1 cursor-pointer accent-primary disabled:cursor-default"
      />
      <span className="w-9 text-[10px] text-muted-foreground sm:w-10 sm:text-[11px]">
        {clock(boundedDuration)}
      </span>
    </div>
  );
}
