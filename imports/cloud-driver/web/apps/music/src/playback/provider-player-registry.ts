import type { MusicClient } from "@cloud-drive/shared";
import type { PlaybackAdapter, PlaybackAdapterEvents } from "./playback-types";

type ProviderPlayerFactory = (
  client: MusicClient,
  events: PlaybackAdapterEvents,
) => PlaybackAdapter;

const providerPlayerFactories = new Map<string, ProviderPlayerFactory>();

export function registerProviderPlayer(
  sdkId: string,
  factory: ProviderPlayerFactory,
): void {
  if (!sdkId.trim()) throw new Error("Provider player SDK ID is required.");
  providerPlayerFactories.set(sdkId, factory);
}

export function createProviderPlayer(
  sdkId: string,
  client: MusicClient,
  events: PlaybackAdapterEvents,
): PlaybackAdapter | undefined {
  return providerPlayerFactories.get(sdkId)?.(client, events);
}
