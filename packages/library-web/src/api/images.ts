import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  CreateImageAlbumRequestSchema,
  CreateImageLinkRequestSchema,
  DeleteImageRequestSchema,
  EmptyImageTrashRequestSchema,
  GetImageRequestSchema,
  ImageService,
  ImportImageRequestSchema,
  ListImageAlbumsRequestSchema,
  ListImageLinksRequestSchema,
  ListImagesRequestSchema,
  PurgeImageRequestSchema,
  RestoreImageRequestSchema,
  RevokeImageLinkRequestSchema,
  UpdateImageRequestSchema,
} from "@realtime-me/library-contracts";
import type { Image, ImageAlbum, ImageLink } from "@realtime-me/library-contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export type ImageListOptions = Readonly<{
  query?: string;
  albumUid?: string;
  trashed?: boolean;
  pageToken?: string;
}>;

export type ImagePage = Readonly<{
  images: Image[];
  nextPageToken: string;
}>;

export class ImagesClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof ImageService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(ImageService, privateTransport(baseUrl));
  }
  async listPage(
    options: ImageListOptions = {},
    signal?: AbortSignal,
  ): Promise<ImagePage> {
    const response = await this.client.listImages(
      create(ListImagesRequestSchema, { ...options, pageSize: 60 }),
      signal ? { signal } : undefined,
    );
    return { images: response.images, nextPageToken: response.nextPageToken };
  }
  async get(imageUid: string): Promise<Image> {
    return required(
      (await this.client.getImage(create(GetImageRequestSchema, { imageUid })))
        .image,
      "image",
    );
  }
  async importUpload(uploadUid: string, albumUid = ""): Promise<Image> {
    return required(
      (
        await this.client.importImage(
          create(ImportImageRequestSchema, { uploadUid, albumUid }),
        )
      ).image,
      "image",
    );
  }
  async update(image: Image): Promise<Image> {
    return required(
      (
        await this.client.updateImage(
          create(UpdateImageRequestSchema, {
            imageUid: image.uid,
            displayName: image.displayName,
            albumUid: image.albumUid,
          }),
        )
      ).image,
      "image",
    );
  }
  async trash(imageUid: string): Promise<void> {
    await this.client.deleteImage(
      create(DeleteImageRequestSchema, { imageUid }),
    );
  }
  async restore(imageUid: string): Promise<void> {
    await this.client.restoreImage(
      create(RestoreImageRequestSchema, { imageUid }),
    );
  }
  async purge(imageUid: string): Promise<void> {
    await this.client.purgeImage(create(PurgeImageRequestSchema, { imageUid }));
  }
  async emptyTrash(): Promise<void> {
    await this.client.emptyImageTrash(create(EmptyImageTrashRequestSchema));
  }
  async albums(signal?: AbortSignal): Promise<ImageAlbum[]> {
    return (
      await this.client.listImageAlbums(
        create(ListImageAlbumsRequestSchema),
        signal ? { signal } : undefined,
      )
    ).albums;
  }
  async createAlbum(displayName: string): Promise<ImageAlbum> {
    return required(
      (
        await this.client.createImageAlbum(
          create(CreateImageAlbumRequestSchema, { displayName }),
        )
      ).album,
      "album",
    );
  }
  async links(imageUid: string): Promise<ImageLink[]> {
    return (
      await this.client.listImageLinks(
        create(ListImageLinksRequestSchema, { imageUid }),
      )
    ).imageLinks;
  }
  async createLink(imageUid: string): Promise<ImageLink> {
    return required(
      (
        await this.client.createImageLink(
          create(CreateImageLinkRequestSchema, { imageUid }),
        )
      ).imageLink,
      "imageLink",
    );
  }
  async revokeLink(imageLinkUid: string): Promise<void> {
    await this.client.revokeImageLink(
      create(RevokeImageLinkRequestSchema, { imageLinkUid }),
    );
  }
  previewUrl(image: Image): string {
    return resolveApiUrl(this.baseUrl, image.previewUrl);
  }
  originalUrl(image: Image): string {
    return resolveApiUrl(this.baseUrl, image.originalUrl);
  }
}
