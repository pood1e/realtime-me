import { Volume2, VolumeX } from "lucide-react";
import { Button } from "@cloud-drive/shared";

export function VolumeControl({
  volume,
  muted,
  onVolumeChange,
  onToggleMuted,
}: {
  volume: number;
  muted: boolean;
  onVolumeChange: (volume: number) => void;
  onToggleMuted: () => void;
}) {
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon-sm"
        onClick={onToggleMuted}
        aria-label={muted ? "取消静音" : "静音"}
      >
        {muted || volume === 0 ? <VolumeX /> : <Volume2 />}
      </Button>
      <input
        aria-label="音量"
        type="range"
        min={0}
        max={1}
        step={0.01}
        value={muted ? 0 : volume}
        onChange={(event) => onVolumeChange(Number(event.target.value))}
        className="hidden w-20 cursor-pointer accent-primary lg:block xl:w-24"
      />
    </div>
  );
}
