import { Footprints, HeartPulse, Music } from 'lucide-react';
import type { PublicStatus } from '@/gen/realtime/me/v1/status_pb';
import { WristState } from '@/gen/realtime/me/v1/watch_pb';

export function Presence({ status }: { status?: PublicStatus | null }) {
  const watch = status?.mobile?.watch;
  const onWrist = watch?.watchState?.wristState !== WristState.OFF_WRIST;
  const heartRate = onWrist ? watch?.heartRate?.beatsPerMinute : undefined;
  const steps = watch?.activityTotals?.steps;
  const media = nowPlaying(status);

  if (!heartRate && !steps && !media) return null;

  return (
    <div className="hidden items-center gap-3.5 text-xs text-muted-foreground md:flex">
      {!!heartRate && (
        <span className="flex items-center gap-1" title="Heart rate">
          <HeartPulse className="size-3.5 text-primary" />
          {heartRate}
        </span>
      )}
      {!!steps && (
        <span className="flex items-center gap-1" title="Steps today">
          <Footprints className="size-3.5" />
          {steps.toLocaleString()}
        </span>
      )}
      {media && (
        <span className="flex max-w-[15rem] items-center gap-1.5 truncate" title={`Now playing: ${media}`}>
          <Music className="size-3.5 text-primary" />
          <span className="truncate">{media}</span>
        </span>
      )}
    </div>
  );
}

function nowPlaying(status?: PublicStatus | null): string | null {
  if (!status) return null;
  const devices = [status.server, ...(status.devices ?? [])];
  for (const device of devices) {
    const media = device?.media;
    if (media?.title) return media.artist ? `${media.title} · ${media.artist}` : media.title;
  }
  return null;
}
