import type { MusicClient } from "@cloud-drive/shared";
import type {
  PlaybackAdapter,
  PlaybackAdapterEvents,
} from "./playback/playback-types";
import { registerProviderPlayer } from "./playback/provider-player-registry";

const SPOTIFY_PROVIDER_ID = "spotify";
const SPOTIFY_SDK_ID = "spotify_web_playback";
const DEVICE_READY_TIMEOUT_MS = 15_000;
const PLAYBACK_REQUEST_TIMEOUT_MS = 15_000;

interface SpotifySDKState {
  paused: boolean;
  position: number;
  duration: number;
}

interface SpotifySDKPlayer {
  addListener(event: string, listener: (value: never) => void): boolean;
  connect(): Promise<boolean>;
  disconnect(): void;
  activateElement(): Promise<void>;
  resume(): Promise<void>;
  pause(): Promise<void>;
  seek(position: number): Promise<void>;
  setVolume(volume: number): Promise<void>;
  getCurrentState(): Promise<SpotifySDKState | null>;
}

interface SpotifySDK {
  Player: new (options: {
    name: string;
    getOAuthToken: (callback: (token: string) => void) => void;
    volume: number;
  }) => SpotifySDKPlayer;
}

declare global {
  interface Window {
    Spotify?: SpotifySDK;
    onSpotifyWebPlaybackSDKReady?: () => void;
  }
}

let sdkPromise: Promise<SpotifySDK> | undefined;

export function registerSpotifyBrowserPlayer(): void {
  registerProviderPlayer(
    SPOTIFY_SDK_ID,
    (client, events) => new SpotifyController(client, events),
  );
}

export class SpotifyController implements PlaybackAdapter {
  private player: SpotifySDKPlayer | undefined;
  private deviceID = "";
  private stateTimer: number | undefined;
  private endTimer: number | undefined;
  private request: AbortController | undefined;
  private volume = 0.8;
  private ended = false;

  constructor(
    private readonly client: MusicClient,
    private readonly events: PlaybackAdapterEvents,
  ) {}

  async load(uri: string): Promise<void> {
    this.clearEndTimer();
    const player = await this.ensurePlayer();
    await player.activateElement();
    this.request?.abort();
    this.request = new AbortController();
    const controller = this.request;
    const timeout = window.setTimeout(
      () => controller.abort(),
      PLAYBACK_REQUEST_TIMEOUT_MS,
    );
    try {
      const token = (
        await this.client.providers.playbackToken(
          SPOTIFY_PROVIDER_ID,
          controller.signal,
        )
      ).accessToken;
      const response = await fetch(
        `https://api.spotify.com/v1/me/player/play?device_id=${encodeURIComponent(this.deviceID)}`,
        {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ uris: [uri] }),
          signal: controller.signal,
        },
      );
      if (!response.ok) throw new Error(spotifyPlaybackError(response.status));
      this.ended = false;
      this.startStateTimer();
    } finally {
      window.clearTimeout(timeout);
      if (this.request === controller) this.request = undefined;
    }
  }

  async play(): Promise<void> {
    await (await this.ensurePlayer()).resume();
  }

  async pause(): Promise<void> {
    await (await this.ensurePlayer()).pause();
    this.clearEndTimer();
  }

  async seek(seconds: number): Promise<void> {
    this.clearEndTimer();
    await (await this.ensurePlayer()).seek(Math.max(0, seconds * 1_000));
  }

  async setVolume(volume: number): Promise<void> {
    this.volume = Math.min(1, Math.max(0, volume));
    if (this.player) await this.player.setVolume(this.volume);
  }

  destroy(): void {
    this.request?.abort();
    this.request = undefined;
    if (this.stateTimer !== undefined) window.clearInterval(this.stateTimer);
    this.stateTimer = undefined;
    this.clearEndTimer();
    this.player?.disconnect();
    this.player = undefined;
    this.deviceID = "";
    this.ended = false;
  }

  private async ensurePlayer(): Promise<SpotifySDKPlayer> {
    if (this.player && this.deviceID) return this.player;
    const sdk = await loadSpotifySDK();
    const player = new sdk.Player({
      name: "Local Library 音乐盒",
      volume: this.volume,
      getOAuthToken: (callback) => {
        void this.client.providers
          .playbackToken(SPOTIFY_PROVIDER_ID)
          .then((token) => callback(token.accessToken))
          .catch((error: unknown) => this.events.onError(error));
      },
    });
    this.player = player;
    player.addListener("playback_state_changed", (state) => {
      if (state) this.publishState(state as SpotifySDKState);
    });
    for (const event of [
      "initialization_error",
      "authentication_error",
      "account_error",
      "playback_error",
      "autoplay_failed",
    ]) {
      player.addListener(event, (value) =>
        this.events.onError(new Error(sdkError(value))),
      );
    }
    const ready = new Promise<string>((resolve, reject) => {
      player.addListener("ready", (value) =>
        resolve((value as { device_id: string }).device_id),
      );
      player.addListener("not_ready", () =>
        reject(new Error("Spotify 播放设备已离线")),
      );
    });
    try {
      if (!(await player.connect())) throw new Error("Spotify 播放器连接失败");
      if (this.player !== player)
        throw new DOMException("Aborted", "AbortError");
      this.deviceID = await withTimeout(
        ready,
        DEVICE_READY_TIMEOUT_MS,
        "Spotify 播放设备连接超时",
      );
      if (this.player !== player)
        throw new DOMException("Aborted", "AbortError");
      return player;
    } catch (error) {
      player.disconnect();
      if (this.player === player) {
        this.player = undefined;
        this.deviceID = "";
      }
      throw error;
    }
  }

  private startStateTimer() {
    if (this.stateTimer !== undefined) return;
    this.stateTimer = window.setInterval(() => {
      const player = this.player;
      if (!player) return;
      void player
        .getCurrentState()
        .then((state) => {
          if (state && this.player === player) this.publishState(state);
        })
        .catch((error: unknown) => {
          if (this.player !== player) return;
          window.clearInterval(this.stateTimer);
          this.stateTimer = undefined;
          this.events.onError(error);
        });
    }, 1_000);
  }

  private publishState(state: SpotifySDKState) {
    const normalized = {
      paused: state.paused,
      position: state.position / 1_000,
      duration: state.duration / 1_000,
    };
    this.events.onState(normalized);
    this.scheduleEnded(normalized);
  }

  private scheduleEnded(state: SpotifySDKState) {
    this.clearEndTimer();
    if (this.ended || state.paused || state.duration <= 0) return;
    const remaining = Math.max(0, state.duration - state.position);
    this.endTimer = window.setTimeout(
      () => {
        this.endTimer = undefined;
        if (this.ended || !this.player) return;
        this.ended = true;
        this.events.onEnded();
      },
      remaining * 1_000 + 300,
    );
  }

  private clearEndTimer() {
    if (this.endTimer !== undefined) window.clearTimeout(this.endTimer);
    this.endTimer = undefined;
  }
}

function loadSpotifySDK(): Promise<SpotifySDK> {
  if (window.Spotify) return Promise.resolve(window.Spotify);
  if (sdkPromise) return sdkPromise;
  sdkPromise = new Promise<SpotifySDK>((resolve, reject) => {
    window.onSpotifyWebPlaybackSDKReady = () => {
      if (window.Spotify) resolve(window.Spotify);
      else {
        sdkPromise = undefined;
        reject(new Error("Spotify SDK 初始化失败"));
      }
    };
    const script = document.createElement("script");
    script.src = "https://sdk.scdn.co/spotify-player.js";
    script.async = true;
    script.onerror = () => {
      sdkPromise = undefined;
      reject(new Error("Spotify SDK 加载失败"));
    };
    document.head.append(script);
  });
  return sdkPromise;
}

function withTimeout<T>(
  operation: Promise<T>,
  milliseconds: number,
  message: string,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const timeout = window.setTimeout(
      () => reject(new Error(message)),
      milliseconds,
    );
    void operation.then(
      (value) => {
        window.clearTimeout(timeout);
        resolve(value);
      },
      (error: unknown) => {
        window.clearTimeout(timeout);
        reject(error);
      },
    );
  });
}

function spotifyPlaybackError(status: number): string {
  if (status === 401) return "Spotify 登录已失效，请重新连接";
  if (status === 403) return "Spotify 播放需要有效的 Premium 账号";
  if (status === 404) return "Spotify 播放设备尚未就绪";
  return `Spotify 播放失败（HTTP ${status}）`;
}

function sdkError(value: unknown): string {
  if (value && typeof value === "object" && "message" in value)
    return String(value.message);
  return "Spotify 播放器发生错误";
}
