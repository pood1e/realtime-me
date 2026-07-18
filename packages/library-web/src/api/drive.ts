import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  CreateDirectoryRequestSchema,
  CreateShareLinkRequestSchema,
  DeleteDriveItemRequestSchema,
  DriveService,
  EmptyTrashRequestSchema,
  GetDownloadRequestSchema,
  ImportDriveFileRequestSchema,
  ListDriveItemsRequestSchema,
  ListShareLinksRequestSchema,
  ListTrashedItemsRequestSchema,
  MoveDriveItemRequestSchema,
  PurgeDriveItemRequestSchema,
  RenameDriveItemRequestSchema,
  RestoreDriveItemRequestSchema,
  RevokeShareLinkRequestSchema,
  SearchDriveItemsRequestSchema,
  ShareService,
} from "@realtime-me/library-contracts";
import type { DriveItem, ShareLink } from "@realtime-me/library-contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export type DriveItemPage = Readonly<{
  items: DriveItem[];
  nextPageToken: string;
}>;

export class DriveClient {
  readonly baseUrl: string;
  private readonly drive: Client<typeof DriveService>;
  private readonly share: Client<typeof ShareService>;

  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    const transport = privateTransport(baseUrl);
    this.drive = createClient(DriveService, transport);
    this.share = createClient(ShareService, transport);
  }

  async listPage(
    parentUid = "",
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<DriveItemPage> {
    const response = await this.drive.listDriveItems(
      create(ListDriveItemsRequestSchema, {
        parentUid,
        pageSize: 100,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return { items: response.items, nextPageToken: response.nextPageToken };
  }
  async listTrashPage(
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<DriveItemPage> {
    const response = await this.drive.listTrashedItems(
      create(ListTrashedItemsRequestSchema, {
        pageSize: 100,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return { items: response.items, nextPageToken: response.nextPageToken };
  }
  async searchPage(
    query: string,
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<DriveItemPage> {
    const response = await this.drive.searchDriveItems(
      create(SearchDriveItemsRequestSchema, {
        query,
        pageSize: 100,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return { items: response.items, nextPageToken: response.nextPageToken };
  }
  async createDirectory(parentUid: string, name: string): Promise<DriveItem> {
    return required(
      (
        await this.drive.createDirectory(
          create(CreateDirectoryRequestSchema, { parentUid, name }),
        )
      ).item,
      "item",
    );
  }
  async rename(itemUid: string, name: string): Promise<DriveItem> {
    return required(
      (
        await this.drive.renameDriveItem(
          create(RenameDriveItemRequestSchema, { itemUid, name }),
        )
      ).item,
      "item",
    );
  }
  async move(itemUid: string, parentUid: string): Promise<DriveItem> {
    return required(
      (
        await this.drive.moveDriveItem(
          create(MoveDriveItemRequestSchema, { itemUid, parentUid }),
        )
      ).item,
      "item",
    );
  }
  async trash(itemUid: string): Promise<DriveItem> {
    return required(
      (
        await this.drive.deleteDriveItem(
          create(DeleteDriveItemRequestSchema, { itemUid }),
        )
      ).item,
      "item",
    );
  }
  async restore(itemUid: string): Promise<DriveItem> {
    return required(
      (
        await this.drive.restoreDriveItem(
          create(RestoreDriveItemRequestSchema, { itemUid }),
        )
      ).item,
      "item",
    );
  }
  async purge(itemUid: string): Promise<void> {
    await this.drive.purgeDriveItem(
      create(PurgeDriveItemRequestSchema, { itemUid }),
    );
  }
  async emptyTrash(): Promise<void> {
    await this.drive.emptyTrash(create(EmptyTrashRequestSchema));
  }
  async importUpload(
    uploadUid: string,
    parentUid: string,
    name: string,
  ): Promise<DriveItem> {
    return required(
      (
        await this.drive.importDriveFile(
          create(ImportDriveFileRequestSchema, { uploadUid, parentUid, name }),
        )
      ).item,
      "item",
    );
  }
  async downloadUrl(itemUid: string): Promise<string> {
    const response = await this.drive.getDownload(
      create(GetDownloadRequestSchema, { itemUid }),
    );
    return resolveApiUrl(this.baseUrl, response.downloadUrl);
  }
  contentUrl(itemUid: string): string {
    return resolveApiUrl(
      this.baseUrl,
      `/v1/items/${encodeURIComponent(itemUid)}/content`,
    );
  }
  async createShare(
    targetUid: string,
    expireTime: Date,
  ): Promise<{ shareLink: ShareLink; shareUrl: string }> {
    const milliseconds = expireTime.getTime();
    const response = await this.share.createShareLink(
      create(CreateShareLinkRequestSchema, {
        targetUid,
        expireTime: {
          seconds: BigInt(Math.floor(milliseconds / 1000)),
          nanos: (milliseconds % 1000) * 1_000_000,
        },
      }),
    );
    return {
      shareLink: required(response.shareLink, "shareLink"),
      shareUrl: response.shareUrl,
    };
  }
  async listShares(targetUid: string): Promise<ShareLink[]> {
    return (
      await this.share.listShareLinks(
        create(ListShareLinksRequestSchema, { targetUid, pageSize: 200 }),
      )
    ).shareLinks;
  }
  async revokeShare(shareUid: string): Promise<ShareLink> {
    return required(
      (
        await this.share.revokeShareLink(
          create(RevokeShareLinkRequestSchema, { shareUid }),
        )
      ).shareLink,
      "shareLink",
    );
  }
}
