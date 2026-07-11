import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  BeginProviderConnectionRequestSchema,
  DeleteTrackRequestSchema,
  DisconnectProviderRequestSchema,
  EmptyTrackTrashRequestSchema,
  GetProviderConnectionAttemptRequestSchema,
  GetProviderLyricsRequestSchema,
  GetSpotifyPlaybackTokenRequestSchema,
  GetTrackRequestSchema,
  ImportTrackRequestSchema,
  ListAlbumsRequestSchema,
  ListArtistsRequestSchema,
  ListPlaybackHistoryRequestSchema,
  ListProviderConnectionsRequestSchema,
  ListTracksRequestSchema,
  MusicService,
  PlaybackQuality,
  ProviderSearchCursorSchema,
  PurgeTrackRequestSchema,
  RecordPlaybackRequestSchema,
  ResolvePlaybackRequestSchema,
  RestoreTrackRequestSchema,
  SearchMusicRequestSchema,
  SetTrackFavoriteRequestSchema,
} from "@cloud-drive/contracts";
import type {
  Album,
  Artist,
  Lyric,
  MusicProvider,
  PlayableTrack,
  PlaybackDescriptor,
  PlaybackEntry,
  ProviderConnection,
  ProviderConnectionAttempt,
  ProviderSearchGroup,
  Track,
} from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export class MusicClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof MusicService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(MusicService, privateTransport(baseUrl));
  }
  async tracks(
    options: {
      query?: string;
      album?: string;
      artist?: string;
      favorites?: boolean;
      trashed?: boolean;
    } = {},
  ): Promise<Track[]> {
    return (
      await this.client.listTracks(
        create(ListTracksRequestSchema, { ...options, pageSize: 200 }),
      )
    ).tracks;
  }
  async get(trackUid: string): Promise<Track> {
    return required(
      (await this.client.getTrack(create(GetTrackRequestSchema, { trackUid })))
        .track,
      "track",
    );
  }
  async importUpload(uploadUid: string): Promise<Track> {
    return required(
      (
        await this.client.importTrack(
          create(ImportTrackRequestSchema, { uploadUid }),
        )
      ).track,
      "track",
    );
  }
  async favorite(trackUid: string, favorite: boolean): Promise<Track> {
    return required(
      (
        await this.client.setTrackFavorite(
          create(SetTrackFavoriteRequestSchema, { trackUid, favorite }),
        )
      ).track,
      "track",
    );
  }
  async trash(trackUid: string): Promise<void> {
    await this.client.deleteTrack(
      create(DeleteTrackRequestSchema, { trackUid }),
    );
  }
  async restore(trackUid: string): Promise<void> {
    await this.client.restoreTrack(
      create(RestoreTrackRequestSchema, { trackUid }),
    );
  }
  async purge(trackUid: string): Promise<void> {
    await this.client.purgeTrack(create(PurgeTrackRequestSchema, { trackUid }));
  }
  async emptyTrash(): Promise<void> {
    await this.client.emptyTrackTrash(create(EmptyTrackTrashRequestSchema));
  }
  async albums(query = ""): Promise<Album[]> {
    return (
      await this.client.listAlbums(create(ListAlbumsRequestSchema, { query }))
    ).albums;
  }
  async artists(query = ""): Promise<Artist[]> {
    return (
      await this.client.listArtists(create(ListArtistsRequestSchema, { query }))
    ).artists;
  }
  async providerConnections(): Promise<ProviderConnection[]> {
    return (
      await this.client.listProviderConnections(
        create(ListProviderConnectionsRequestSchema),
      )
    ).connections;
  }
  async beginProviderConnection(
    provider: MusicProvider,
  ): Promise<ProviderConnectionAttempt> {
    return required(
      (
        await this.client.beginProviderConnection(
          create(BeginProviderConnectionRequestSchema, { provider }),
        )
      ).attempt,
      "provider connection attempt",
    );
  }
  async providerConnectionAttempt(
    attemptUid: string,
  ): Promise<ProviderConnectionAttempt> {
    return required(
      (
        await this.client.getProviderConnectionAttempt(
          create(GetProviderConnectionAttemptRequestSchema, { attemptUid }),
        )
      ).attempt,
      "provider connection attempt",
    );
  }
  async disconnectProvider(provider: MusicProvider): Promise<void> {
    await this.client.disconnectProvider(
      create(DisconnectProviderRequestSchema, { provider }),
    );
  }
  async searchMusic(
    query: string,
    cursors: Array<{ provider: MusicProvider; pageToken?: string }> = [],
  ): Promise<ProviderSearchGroup[]> {
    const response = await this.client.searchMusic(
      create(SearchMusicRequestSchema, {
        query,
        cursors: cursors.map((cursor) =>
          create(ProviderSearchCursorSchema, cursor),
        ),
      }),
    );
    return response.groups;
  }
  async resolvePlayback(
    track: PlayableTrack,
    quality = PlaybackQuality.BEST_COMPATIBLE,
  ): Promise<PlaybackDescriptor> {
    return required(
      (
        await this.client.resolvePlayback(
          create(ResolvePlaybackRequestSchema, {
            provider: track.provider,
            trackId: track.trackId,
            quality,
          }),
        )
      ).playback,
      "playback descriptor",
    );
  }
  async providerLyrics(track: PlayableTrack): Promise<Lyric> {
    return required(
      (
        await this.client.getProviderLyrics(
          create(GetProviderLyricsRequestSchema, {
            provider: track.provider,
            trackId: track.trackId,
          }),
        )
      ).lyric,
      "lyrics",
    );
  }
  async spotifyPlaybackToken() {
    return this.client.getSpotifyPlaybackToken(
      create(GetSpotifyPlaybackTokenRequestSchema),
    );
  }
  async recordPlayback(track: PlayableTrack): Promise<void> {
    await this.client.recordPlayback(
      create(RecordPlaybackRequestSchema, { track }),
    );
  }
  async history(): Promise<PlaybackEntry[]> {
    return (
      await this.client.listPlaybackHistory(
        create(ListPlaybackHistoryRequestSchema, { pageSize: 100 }),
      )
    ).playbackEntries;
  }
  contentUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.contentUrl);
  }
  artworkUrl(track: Track): string {
    return resolveApiUrl(this.baseUrl, track.artworkUrl);
  }
  playableArtworkUrl(track: PlayableTrack): string {
    return track.artworkUrl
      ? resolveApiUrl(this.baseUrl, track.artworkUrl)
      : "";
  }
  playbackUrl(url: string): string {
    return resolveApiUrl(this.baseUrl, url);
  }
}
