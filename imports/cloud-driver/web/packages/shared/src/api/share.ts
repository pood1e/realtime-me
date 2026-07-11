import { create } from "@bufbuild/protobuf";
import { createClient, type Client } from "@connectrpc/connect";
import {
  ListSharedItemsRequestSchema,
  ResolveShareRequestSchema,
  ShareService,
} from "@cloud-drive/contracts";
import type { DriveItem, ShareLink } from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  publicTransport,
  required,
  resolveApiUrl,
} from "./core";

export type SharedItemPage = Readonly<{
  items: DriveItem[];
  nextPageToken: string;
}>;

export class PublicShareClient {
  private readonly baseUrl: string;
  private readonly client: Client<typeof ShareService>;

  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(ShareService, publicTransport(baseUrl));
  }

  async resolveShare(
    token: string,
    signal?: AbortSignal,
  ): Promise<ResolvedShare> {
    const response = await this.client.resolveShare(
      create(ResolveShareRequestSchema, { shareToken: token }),
      signal ? { signal } : undefined,
    );
    return {
      shareLink: required(response.shareLink, "shareLink"),
      target: required(response.target, "target"),
    };
  }

  async listSharedItemsPage(
    token: string,
    parentUid: string,
    pageToken = "",
    signal?: AbortSignal,
  ): Promise<SharedItemPage> {
    const response = await this.client.listSharedItems(
      create(ListSharedItemsRequestSchema, {
        shareToken: token,
        parentUid,
        pageSize: 100,
        pageToken,
      }),
      signal ? { signal } : undefined,
    );
    return { items: response.items, nextPageToken: response.nextPageToken };
  }

  contentUrl(token: string, itemUid: string): string {
    return resolveApiUrl(
      this.baseUrl,
      `/v1/shares/${encodeURIComponent(token)}/items/${encodeURIComponent(itemUid)}/content`,
    );
  }
}

export type ResolvedShare = Readonly<{
  shareLink: ShareLink;
  target: DriveItem;
}>;
