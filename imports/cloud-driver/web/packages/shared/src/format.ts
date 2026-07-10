import type { DriveItem } from "@cloud-drive/contracts";

import { driveItemContentType, driveItemName } from "./message";

const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"];

export function formatBytes(bytes: number | bigint): string {
  const numericBytes = typeof bytes === "bigint" ? Number(bytes) : bytes;
  if (!Number.isFinite(numericBytes) || numericBytes <= 0) {
    return "0 B";
  }

  const exponent = Math.min(Math.floor(Math.log(numericBytes) / Math.log(1_024)), BYTE_UNITS.length - 1);
  const value = numericBytes / 1_024 ** exponent;
  const precision = value >= 10 || exponent === 0 ? 0 : 1;
  return `${value.toFixed(precision)} ${BYTE_UNITS[exponent]}`;
}

export function formatDate(value: Date | undefined): string {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(value);
}

export function fileExtension(name: string): string {
  const dot = name.lastIndexOf(".");
  return dot <= 0 ? "" : name.slice(dot + 1).toLowerCase();
}

export function isImage(item: DriveItem): boolean {
  const contentType = driveItemContentType(item);
  return contentType.startsWith("image/") || ["avif", "gif", "jpg", "jpeg", "png", "svg", "webp"].includes(fileExtension(driveItemName(item)));
}

export function isPdf(item: DriveItem): boolean {
  return driveItemContentType(item) === "application/pdf" || fileExtension(driveItemName(item)) === "pdf";
}

export function isText(item: DriveItem): boolean {
  const contentType = driveItemContentType(item);
  if (contentType.startsWith("text/")) {
    return true;
  }
  return ["csv", "css", "go", "html", "js", "json", "md", "py", "rs", "sh", "sql", "ts", "tsx", "txt", "xml", "yaml", "yml"].includes(fileExtension(driveItemName(item)));
}
