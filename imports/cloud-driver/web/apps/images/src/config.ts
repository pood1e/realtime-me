import { privateAppConfiguration } from "@cloud-drive/shared";

const configuration = privateAppConfiguration(import.meta.env);

export const API_BASE = configuration.apiBase;
export const AUTH_ORIGIN = configuration.authOrigin;
export const APP_LINKS = configuration.links;
