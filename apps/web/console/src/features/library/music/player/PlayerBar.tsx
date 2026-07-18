import type { MusicClient } from "@realtime-me/library-web";
import { TooltipProvider } from "@realtime-me/web-ui";
import type { PlaybackQueueController } from "../playback/playback-queue";
import type { PlaybackSessionController } from "../playback/use-playback-session";
import { NowPlaying } from "./NowPlaying";
import { PlaybackControls } from "./PlaybackControls";
import { PlaybackProgress } from "./PlaybackProgress";
import { PlaybackQueueSheet } from "./PlaybackQueueSheet";
import { VolumeControl } from "./VolumeControl";

export function PlayerBar({
  client,
  queue,
  session,
  volume,
  muted,
  onVolumeChange,
  onToggleMuted,
}: {
  client: MusicClient;
  queue: PlaybackQueueController;
  session: PlaybackSessionController;
  volume: number;
  muted: boolean;
  onVolumeChange: (volume: number) => void;
  onToggleMuted: () => void;
}) {
  const track = queue.currentTrack;
  if (!track) return null;
  return (
    <TooltipProvider>
      <div className="fixed right-0 bottom-0 left-0 z-40 border-t bg-card/95 px-3 py-2 shadow-2xl backdrop-blur lg:left-60">
        <div className="mx-auto max-w-[92rem]">
          <div className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-2 md:grid-cols-[minmax(12rem,1fr)_auto_minmax(12rem,1fr)]">
            <NowPlaying
              track={track}
              descriptor={session.descriptor}
              client={client}
              error={session.error}
            />
            <div className="col-span-2 row-start-2 md:col-span-1 md:col-start-2 md:row-start-1">
              <PlaybackControls
                mode={queue.mode}
                playing={session.playing}
                loading={session.loading}
                canPrevious={queue.canPrevious}
                canNext={queue.canNext && !queue.loadingMore}
                onCycleMode={queue.cycleMode}
                onPrevious={() => queue.previous(session.position)}
                onToggle={session.toggle}
                onNext={() => void queue.next()}
              />
            </div>
            <div className="col-start-2 row-start-1 flex items-center justify-end gap-1 md:col-start-3">
              <VolumeControl
                volume={volume}
                muted={muted}
                onVolumeChange={onVolumeChange}
                onToggleMuted={onToggleMuted}
              />
              <PlaybackQueueSheet client={client} queue={queue} />
            </div>
          </div>
          <div className="mt-1.5">
            <PlaybackProgress
              position={session.position}
              duration={session.duration}
              onSeek={session.seek}
            />
          </div>
        </div>
      </div>
    </TooltipProvider>
  );
}
