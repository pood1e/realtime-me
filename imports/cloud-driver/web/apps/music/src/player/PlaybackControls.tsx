import {
  ListEnd,
  LoaderCircle,
  Pause,
  Play,
  Repeat1,
  Repeat2,
  Shuffle,
  SkipBack,
  SkipForward,
  type LucideIcon,
} from "lucide-react";
import type { ReactElement } from "react";
import {
  Button,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@cloud-drive/shared";
import type { PlaybackMode } from "../playback/playback-types";

const MODE_PRESENTATION: Record<
  PlaybackMode,
  Readonly<{ label: string; icon: LucideIcon }>
> = {
  sequential: { label: "顺序播放", icon: ListEnd },
  "repeat-all": { label: "列表循环", icon: Repeat2 },
  "repeat-one": { label: "单曲循环", icon: Repeat1 },
  shuffle: { label: "随机播放", icon: Shuffle },
};

export function PlaybackControls({
  mode,
  playing,
  loading,
  canPrevious,
  canNext,
  onCycleMode,
  onPrevious,
  onToggle,
  onNext,
}: {
  mode: PlaybackMode;
  playing: boolean;
  loading: boolean;
  canPrevious: boolean;
  canNext: boolean;
  onCycleMode: () => void;
  onPrevious: () => void;
  onToggle: () => void;
  onNext: () => void;
}) {
  const modePresentation = MODE_PRESENTATION[mode];
  const ModeIcon = modePresentation.icon;
  return (
    <div className="flex items-center justify-center gap-1 sm:gap-2">
      <ControlTip label={modePresentation.label}>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onCycleMode}
          aria-label={modePresentation.label}
        >
          <ModeIcon />
        </Button>
      </ControlTip>
      <ControlTip label="上一首">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onPrevious}
          disabled={!canPrevious}
          aria-label="上一首"
        >
          <SkipBack />
        </Button>
      </ControlTip>
      <ControlTip label={playing ? "暂停" : "播放"}>
        <Button
          size="icon-lg"
          onClick={onToggle}
          disabled={loading}
          aria-label={playing ? "暂停" : "播放"}
        >
          {loading ? (
            <LoaderCircle className="animate-spin" />
          ) : playing ? (
            <Pause />
          ) : (
            <Play />
          )}
        </Button>
      </ControlTip>
      <ControlTip label="下一首">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onNext}
          disabled={!canNext}
          aria-label="下一首"
        >
          <SkipForward />
        </Button>
      </ControlTip>
    </div>
  );
}

function ControlTip({
  label,
  children,
}: {
  label: string;
  children: ReactElement;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>{children}</TooltipTrigger>
      <TooltipContent side="top" sideOffset={8}>
        {label}
      </TooltipContent>
    </Tooltip>
  );
}
