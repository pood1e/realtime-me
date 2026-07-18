import { DriveItemKind } from "@realtime-me/library-contracts";
import type { DriveItem, ShareLink, Upload } from "@realtime-me/library-contracts";

const MAX_SAFE_BIGINT = BigInt(Number.MAX_SAFE_INTEGER);

function timestampToDate(
  value: DriveItem["updateTime"] | ShareLink["expireTime"],
): Date | undefined {
  if (!value) {
    return undefined;
  }

  if (
    value.seconds > MAX_SAFE_BIGINT / 1_000n ||
    value.seconds < -MAX_SAFE_BIGINT / 1_000n
  ) {
    return undefined;
  }

  const date = new Date(
    Number(value.seconds) * 1_000 + Math.trunc(value.nanos / 1_000_000),
  );
  return Number.isNaN(date.getTime()) ? undefined : date;
}

function asByteOffset(value: bigint): number {
  if (value < 0n || value > MAX_SAFE_BIGINT) {
    throw new RangeError(
      "The browser cannot address this upload range safely.",
    );
  }
  return Number(value);
}

export function driveItemUid(item: DriveItem): string {
  return item.uid;
}

export function driveItemParentUid(item: DriveItem): string {
  return item.parentUid;
}

export function driveItemName(item: DriveItem): string {
  return item.name || "Untitled";
}

export function driveItemContentType(item: DriveItem): string {
  return item.contentType;
}

export function driveItemSize(item: DriveItem): bigint {
  return item.sizeBytes;
}

export function driveItemUpdatedAt(item: DriveItem): Date | undefined {
  return timestampToDate(item.updateTime);
}

export function driveItemDeletedAt(item: DriveItem): Date | undefined {
  return timestampToDate(item.deleteTime);
}

export function driveItemIsDirectory(item: DriveItem): boolean {
  return item.kind === DriveItemKind.DIRECTORY;
}

export function shareLinkUid(shareLink: ShareLink): string {
  return shareLink.uid;
}

export function shareLinkExpiresAt(shareLink: ShareLink): Date | undefined {
  return timestampToDate(shareLink.expireTime);
}

export function shareLinkRevokedAt(shareLink: ShareLink): Date | undefined {
  return timestampToDate(shareLink.revokeTime);
}

export function uploadUid(upload: Upload): string {
  return upload.uid;
}

export function uploadChunkSize(upload: Upload): number {
  return asByteOffset(upload.chunkSizeBytes);
}

export function uploadRanges(
  upload: Upload,
): ReadonlyArray<readonly [number, number]> {
  return upload.chunks.map(
    (chunk) =>
      [asByteOffset(chunk.startOffset), asByteOffset(chunk.endOffset)] as const,
  );
}

export function uploadReceivedBytes(upload: Upload): number {
  return asByteOffset(upload.receivedBytes);
}
