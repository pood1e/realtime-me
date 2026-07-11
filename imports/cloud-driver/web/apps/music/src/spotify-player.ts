import type { MusicClient } from "@cloud-drive/shared";

export interface SpotifyPlayerState {
  paused: boolean;
  position: number;
  duration: number;
}

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
  togglePlay(): Promise<void>;
  seek(position: number): Promise<void>;
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

export class SpotifyController {
  private player?: SpotifySDKPlayer;
  private deviceID = "";
  private stateTimer?: number;

  constructor(
    private readonly client: MusicClient,
    private readonly onState: (state: SpotifyPlayerState) => void,
    private readonly onError: (message: string) => void,
  ) {}

  async play(uri: string): Promise<void> {
    const player = await this.ensurePlayer();
    await player.activateElement();
    const token = (await this.client.spotifyPlaybackToken()).accessToken;
    const response = await fetch(
      `https://api.spotify.com/v1/me/player/play?device_id=${encodeURIComponent(this.deviceID)}`,
      {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ uris: [uri] }),
      },
    );
    if (!response.ok) throw new Error(spotifyPlaybackError(response.status));
    this.startStateTimer();
  }

  async toggle(): Promise<void> {
    await (await this.ensurePlayer()).togglePlay();
  }

  async seek(seconds: number): Promise<void> {
    await (await this.ensurePlayer()).seek(Math.max(0, seconds * 1_000));
  }

  disconnect(): void {
    if (this.stateTimer) window.clearInterval(this.stateTimer);
    this.stateTimer = undefined;
    this.player?.disconnect();
    this.player = undefined;
    this.deviceID = "";
  }

  private async ensurePlayer(): Promise<SpotifySDKPlayer> {
    if (this.player && this.deviceID) return this.player;
    const sdk = await loadSpotifySDK();
    const player = new sdk.Player({
      name: "Local Library 音乐盒",
      volume: 0.8,
      getOAuthToken: (callback) => {
        void this.client
          .spotifyPlaybackToken()
          .then((token) => callback(token.accessToken))
          .catch((error: unknown) => this.onError(message(error)));
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
      player.addListener(event, (value) => this.onError(sdkError(value)));
    }
    const ready = new Promise<string>((resolve, reject) => {
      player.addListener("ready", (value) =>
        resolve((value as { device_id: string }).device_id),
      );
      player.addListener("not_ready", () =>
        reject(new Error("Spotify 播放设备已离线")),
      );
    });
    if (!(await player.connect())) throw new Error("Spotify 播放器连接失败");
    this.deviceID = await ready;
    return player;
  }

  private startStateTimer() {
    if (this.stateTimer) return;
    this.stateTimer = window.setInterval(() => {
      void this.player?.getCurrentState().then((state) => {
        if (state) this.publishState(state);
      });
    }, 1_000);
  }

  private publishState(state: SpotifySDKState) {
    this.onState({
      paused: state.paused,
      position: state.position / 1_000,
      duration: state.duration / 1_000,
    });
  }
}

function loadSpotifySDK(): Promise<SpotifySDK> {
  if (window.Spotify) return Promise.resolve(window.Spotify);
  if (sdkPromise) return sdkPromise;
  sdkPromise = new Promise<SpotifySDK>((resolve, reject) => {
    window.onSpotifyWebPlaybackSDKReady = () => {
      if (window.Spotify) resolve(window.Spotify);
      else reject(new Error("Spotify SDK 初始化失败"));
    };
    const script = document.createElement("script");
    script.src = "https://sdk.scdn.co/spotify-player.js";
    script.async = true;
    script.onerror = () => reject(new Error("Spotify SDK 加载失败"));
    document.head.append(script);
  });
  return sdkPromise;
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

function message(error: unknown): string {
  return error instanceof Error ? error.message : "Spotify 授权失败";
}
