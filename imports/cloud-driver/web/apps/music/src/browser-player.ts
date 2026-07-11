import type { MusicClient } from "@cloud-drive/shared";

export type BrowserPlayerState = Readonly<{
  paused: boolean;
  position: number;
  duration: number;
}>;

export interface BrowserPlayer {
  play(resourceUri: string): Promise<void>;
  toggle(): Promise<void>;
  seek(seconds: number): Promise<void>;
  disconnect(): void;
}

type BrowserPlayerFactory = (
  client: MusicClient,
  onState: (state: BrowserPlayerState) => void,
  onError: (message: string) => void,
) => BrowserPlayer;

const browserPlayerFactories = new Map<string, BrowserPlayerFactory>();

export function registerBrowserPlayer(
  sdkId: string,
  factory: BrowserPlayerFactory,
): void {
  if (!sdkId.trim()) throw new Error("Browser player SDK ID is required.");
  browserPlayerFactories.set(sdkId, factory);
}

export function createBrowserPlayer(
  sdkId: string,
  client: MusicClient,
  onState: (state: BrowserPlayerState) => void,
  onError: (message: string) => void,
): BrowserPlayer | undefined {
  return browserPlayerFactories.get(sdkId)?.(client, onState, onError);
}
