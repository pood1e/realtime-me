import { create } from "@bufbuild/protobuf";
import { createClient, type Client, type Transport } from "@connectrpc/connect";
import {
  DeletePlaylistRequestSchema,
  DownloadPlaylistRequestSchema,
  GetPlaylistImportRequestSchema,
  GetPlaylistRequestSchema,
  ImportPlaylistRequestSchema,
  ListPlaylistsRequestSchema,
  ListPlaylistTracksRequestSchema,
  MusicPlaylistService,
  PlaylistImportStatus,
} from "@cloud-drive/contracts";
import type { Playlist, PlaylistTrack } from "@cloud-drive/contracts";

import { privateTransport, required } from "./core";
import { abortableDelay } from "./delay";
import type { ProviderId } from "./music-providers";

const PLAYLIST_IMPORT_TIMEOUT_MS = 15 * 60_000;

export type PlaylistPage = Readonly<{
  playlists: Playlist[];
  nextPageToken: string;
}>;

export type PlaylistTrackPage = Readonly<{
  tracks: PlaylistTrack[];
  nextPageToken: string;
}>;

export class MusicPlaylistClient {
  private readonly client: Client<typeof MusicPlaylistService>;

  constructor(
    baseUrl: string,
    transport: Transport = privateTransport(baseUrl),
  ) {
    this.client = createClient(MusicPlaylistService, transport);
  }

  async importPlaylist(
    providerId: ProviderId,
    source: string,
    signal?: AbortSignal,
  ): Promise<Playlist> {
    let operation = required(
      (
        await this.client.importPlaylist(
          create(ImportPlaylistRequestSchema, { providerId, source }),
          signal ? { signal } : undefined,
        )
      ).playlistImport,
      "playlist import",
    );
    const deadline = Date.now() + PLAYLIST_IMPORT_TIMEOUT_MS;
    while (
      operation.status === PlaylistImportStatus.PENDING ||
      operation.status === PlaylistImportStatus.RUNNING
    ) {
      if (Date.now() >= deadline) {
        throw new Error("歌单仍在后台导入，请稍后刷新歌单列表");
      }
      await abortableDelay(750, signal);
      operation = required(
        (
          await this.client.getPlaylistImport(
            create(GetPlaylistImportRequestSchema, {
              playlistImportUid: operation.uid,
            }),
            signal ? { signal } : undefined,
          )
        ).playlistImport,
        "playlist import",
      );
    }
    if (
      operation.status !== PlaylistImportStatus.COMPLETED ||
      !operation.playlistUid
    ) {
      throw new Error(
        operation.failureCode
          ? `歌单导入失败（${operation.failureCode}）`
          : "歌单导入失败",
      );
    }
    return this.get(operation.playlistUid, signal);
  }

  async get(playlistUid: string, signal?: AbortSignal): Promise<Playlist> {
    return required(
      (
        await this.client.getPlaylist(
          create(GetPlaylistRequestSchema, { playlistUid }),
          signal ? { signal } : undefined,
        )
      ).playlist,
      "playlist",
    );
  }

  async page(pageToken = "", signal?: AbortSignal): Promise<PlaylistPage> {
    const response = await this.client.listPlaylists(
      create(ListPlaylistsRequestSchema, { pageSize: 50, pageToken }),
      signal ? { signal } : undefined,
    );
    return {
      playlists: response.playlists,
      nextPageToken: response.nextPageToken,
    };
  }

  async tracks(
    playlistUid: string,
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<PlaylistTrackPage> {
    const response = await this.client.listPlaylistTracks(
      create(ListPlaylistTracksRequestSchema, {
        playlistUid,
        pageSize: 100,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return {
      tracks: response.playlistTracks,
      nextPageToken: response.nextPageToken,
    };
  }

  async download(playlistUid: string): Promise<Playlist> {
    return required(
      (
        await this.client.downloadPlaylist(
          create(DownloadPlaylistRequestSchema, { playlistUid }),
        )
      ).playlist,
      "playlist",
    );
  }

  async delete(playlistUid: string): Promise<void> {
    await this.client.deletePlaylist(
      create(DeletePlaylistRequestSchema, { playlistUid }),
    );
  }
}
