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
} from "@cloud-drive/contracts";
import type { Image, ImageAlbum, ImageLink } from "@cloud-drive/contracts";

import {
  normalizeBaseUrl,
  privateTransport,
  required,
  resolveApiUrl,
} from "./core";

export class ImagesClient {
  readonly baseUrl: string;
  private readonly client: Client<typeof ImageService>;
  constructor(baseUrl: string) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.client = createClient(ImageService, privateTransport(baseUrl));
  }
  async list(
    options: { query?: string; albumUid?: string; trashed?: boolean } = {},
  ): Promise<Image[]> {
    return (
      await this.client.listImages(
        create(ListImagesRequestSchema, { ...options, pageSize: 200 }),
      )
    ).images;
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
  async albums(): Promise<ImageAlbum[]> {
    return (
      await this.client.listImageAlbums(create(ListImageAlbumsRequestSchema))
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
