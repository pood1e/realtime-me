import type {
  PlaybackAdapter,
  PlaybackAdapterEvents,
  PlaybackAdapterState,
} from "./playback-types";

export class DirectAudioPlayer implements PlaybackAdapter {
  private readonly audio = new Audio();
  private active = true;

  constructor(
    credentialed: boolean,
    private readonly events: PlaybackAdapterEvents,
  ) {
    if (credentialed) this.audio.crossOrigin = "use-credentials";
    this.audio.preload = "metadata";
    this.audio.addEventListener("play", this.publishState);
    this.audio.addEventListener("pause", this.publishState);
    this.audio.addEventListener("timeupdate", this.publishState);
    this.audio.addEventListener("durationchange", this.publishState);
    this.audio.addEventListener("ended", this.handleEnded);
    this.audio.addEventListener("error", this.handleError);
  }

  async load(resource: string): Promise<void> {
    this.audio.src = resource;
    this.audio.load();
    await this.audio.play();
  }

  async play(): Promise<void> {
    await this.audio.play();
  }

  pause(): Promise<void> {
    this.audio.pause();
    return Promise.resolve();
  }

  seek(seconds: number): Promise<void> {
    this.audio.currentTime = finiteSeconds(seconds);
    this.publishState();
    return Promise.resolve();
  }

  setVolume(volume: number): Promise<void> {
    this.audio.volume = Math.min(1, Math.max(0, volume));
    return Promise.resolve();
  }

  destroy(): void {
    if (!this.active) return;
    this.active = false;
    this.audio.pause();
    this.audio.removeEventListener("play", this.publishState);
    this.audio.removeEventListener("pause", this.publishState);
    this.audio.removeEventListener("timeupdate", this.publishState);
    this.audio.removeEventListener("durationchange", this.publishState);
    this.audio.removeEventListener("ended", this.handleEnded);
    this.audio.removeEventListener("error", this.handleError);
    this.audio.removeAttribute("src");
    this.audio.load();
  }

  private readonly publishState = () => {
    if (!this.active) return;
    this.events.onState(audioState(this.audio));
  };

  private readonly handleEnded = () => {
    if (!this.active) return;
    this.publishState();
    this.events.onEnded();
  };

  private readonly handleError = () => {
    if (!this.active) return;
    this.events.onError(new Error("当前音频无法播放"));
  };
}

function audioState(audio: HTMLAudioElement): PlaybackAdapterState {
  return {
    paused: audio.paused,
    position: finiteSeconds(audio.currentTime),
    duration: finiteSeconds(audio.duration),
  };
}

function finiteSeconds(value: number): number {
  return Number.isFinite(value) ? Math.max(0, value) : 0;
}
