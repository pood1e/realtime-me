function configuredOptionalUrl(value: string | undefined): string | undefined {
  const candidate = value?.trim().replace(/\/+$/, "");
  if (!candidate) return undefined;
  let url: URL;
  try {
    url = new URL(candidate);
  } catch {
    return undefined;
  }
  const loopback =
    url.hostname === "localhost" || url.hostname === "[::1]" || url.hostname.startsWith("127.");
  if (
    (url.protocol !== "https:" && !(url.protocol === "http:" && loopback)) ||
    url.username ||
    url.password ||
    (url.pathname !== "/" && url.pathname !== "") ||
    url.search ||
    url.hash
  ) {
    return undefined;
  }
  return url.origin;
}

export const STATUS_API_BASE = window.location.origin;
export const PUBLIC_LIBRARY_API_BASE = window.location.origin;
export const CONSOLE_URL = configuredOptionalUrl(import.meta.env.VITE_CONSOLE_URL);
