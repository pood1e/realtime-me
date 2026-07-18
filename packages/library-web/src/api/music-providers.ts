import { create } from "@bufbuild/protobuf";
import { type Client, createClient, type Transport } from "@connectrpc/connect";
import type {
  Lyric,
  PlayableTrack,
  PlaybackDescriptor,
  ProviderConnection,
  ProviderConnectionAttempt,
  ProviderDescriptor,
  ProviderSearchGroup,
} from "@realtime-me/library-contracts";
import {
  BeginProviderConnectionRequestSchema,
  DisconnectProviderRequestSchema,
  GetProviderConnectionAttemptRequestSchema,
  GetProviderLyricsRequestSchema,
  GetProviderPlaybackTokenRequestSchema,
  ListProviderConnectionsRequestSchema,
  ListProvidersRequestSchema,
  MusicProviderService,
  PlaybackQuality,
  ProviderSearchCursorSchema,
  ResolvePlaybackRequestSchema,
  SearchMusicRequestSchema,
} from "@realtime-me/library-contracts";

import { normalizeBaseUrl, privateTransport, required, resolveApiUrl } from "./core";

export type ProviderId = string;

export const LOCAL_PROVIDER_ID = "local";

export class MusicProviderClient {
  private readonly baseUrl: string;
  private readonly client: Client<typeof MusicProviderService>;

  constructor(baseUrl: string, transport: Transport = privateTransport(baseUrl)) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(MusicProviderService, transport);
  }

  async descriptors(signal?: AbortSignal): Promise<ProviderDescriptor[]> {
    return (
      await this.client.listProviders(
        create(ListProvidersRequestSchema),
        signal ? { signal } : undefined,
      )
    ).providers;
  }

  async connections(signal?: AbortSignal): Promise<ProviderConnection[]> {
    return (
      await this.client.listProviderConnections(
        create(ListProviderConnectionsRequestSchema),
        signal ? { signal } : undefined,
      )
    ).connections;
  }

  async beginConnection(providerId: ProviderId): Promise<ProviderConnectionAttempt> {
    return required(
      (
        await this.client.beginProviderConnection(
          create(BeginProviderConnectionRequestSchema, { providerId }),
        )
      ).attempt,
      "provider connection attempt",
    );
  }

  async connectionAttempt(
    attemptUid: string,
    signal?: AbortSignal,
  ): Promise<ProviderConnectionAttempt> {
    return required(
      (
        await this.client.getProviderConnectionAttempt(
          create(GetProviderConnectionAttemptRequestSchema, { attemptUid }),
          signal ? { signal } : undefined,
        )
      ).attempt,
      "provider connection attempt",
    );
  }

  async disconnect(providerId: ProviderId): Promise<void> {
    await this.client.disconnectProvider(create(DisconnectProviderRequestSchema, { providerId }));
  }

  async search(
    query: string,
    cursors: Array<{ providerId: ProviderId; pageToken?: string }> = [],
    signal?: AbortSignal,
  ): Promise<ProviderSearchGroup[]> {
    const response = await this.client.searchMusic(
      create(SearchMusicRequestSchema, {
        query,
        cursors: cursors.map((cursor) => create(ProviderSearchCursorSchema, cursor)),
      }),
      signal ? { signal } : undefined,
    );
    return response.groups;
  }

  async resolvePlayback(
    track: PlayableTrack,
    quality = PlaybackQuality.BEST_COMPATIBLE,
    signal?: AbortSignal,
  ): Promise<PlaybackDescriptor> {
    return required(
      (
        await this.client.resolvePlayback(
          create(ResolvePlaybackRequestSchema, {
            providerId: track.providerId,
            trackId: track.trackId,
            quality,
          }),
          signal ? { signal } : undefined,
        )
      ).playback,
      "playback descriptor",
    );
  }

  async lyrics(track: PlayableTrack, signal?: AbortSignal): Promise<Lyric> {
    return required(
      (
        await this.client.getProviderLyrics(
          create(GetProviderLyricsRequestSchema, {
            providerId: track.providerId,
            trackId: track.trackId,
          }),
          signal ? { signal } : undefined,
        )
      ).lyric,
      "lyrics",
    );
  }

  async playbackToken(providerId: ProviderId, signal?: AbortSignal) {
    return this.client.getProviderPlaybackToken(
      create(GetProviderPlaybackTokenRequestSchema, { providerId }),
      signal ? { signal } : undefined,
    );
  }

  artworkUrl(track: PlayableTrack): string {
    return track.artworkUrl ? resolveApiUrl(this.baseUrl, track.artworkUrl) : "";
  }

  playbackUrl(url: string): string {
    return resolveApiUrl(this.baseUrl, url);
  }
}
