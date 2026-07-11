import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  type PropsWithChildren,
} from "react";
import type { ProviderDescriptor } from "@cloud-drive/contracts";
import {
  LOCAL_PROVIDER_ID,
  MusicClient,
  type ProviderId,
  useQuery,
} from "@cloud-drive/shared";

const ProviderNames = createContext<ReadonlyMap<ProviderId, string>>(new Map());

export function MusicProviderCatalog({
  client,
  children,
}: PropsWithChildren<{ client: MusicClient }>) {
  const providers = useQuery({
    queryKey: ["music-provider-descriptors"],
    queryFn: ({ signal }) => client.providers.descriptors(signal),
  });
  const names = useMemo(
    () =>
      new Map(
        (providers.data ?? []).map((provider: ProviderDescriptor) => [
          provider.id,
          provider.displayName,
        ]),
      ),
    [providers.data],
  );
  return (
    <ProviderNames.Provider value={names}>{children}</ProviderNames.Provider>
  );
}

export function useProviderLabel() {
  const names = useContext(ProviderNames);
  return useCallback(
    (providerId: ProviderId) =>
      names.get(providerId) ?? fallbackProviderName(providerId),
    [names],
  );
}

function fallbackProviderName(providerId: ProviderId): string {
  if (providerId === LOCAL_PROVIDER_ID) return "本地音乐";
  return providerId || "未知来源";
}
