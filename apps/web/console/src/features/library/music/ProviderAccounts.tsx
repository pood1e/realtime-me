import {
  type ProviderConnection,
  type ProviderConnectionAttempt,
  ProviderConnectionAttemptStatus,
} from "@realtime-me/library-contracts";
import {
  LoadingIndicator,
  type MusicClient,
  type ProviderId,
  useDialog,
  useQuery,
  useToast,
} from "@realtime-me/library-web";
import { useCallback, useEffect, useState } from "react";
import { ProviderAccountRow } from "./ProviderAccountRow";
import { ProviderLoginDialog } from "./ProviderLoginDialog";
import { terminalConnectionAttempt } from "./provider-account-model";
import { useProviderLabel } from "./provider-catalog";

export function ProviderAccounts({ client }: { client: MusicClient }) {
  const { showToast } = useToast();
  const providerLabel = useProviderLabel();
  const { confirm } = useDialog();
  const [attempt, setAttempt] = useState<ProviderConnectionAttempt>();
  const connections = useQuery({
    queryKey: ["music-provider-connections"],
    queryFn: ({ signal }) => client.providers.connections(signal),
  });
  const reload = useCallback(async () => {
    await connections.refetch();
  }, [connections.refetch]);
  const attemptQuery = useQuery({
    queryKey: ["music-provider-attempt", attempt?.uid ?? ""],
    enabled: Boolean(attempt && !terminalConnectionAttempt(attempt.status)),
    queryFn: ({ signal }) => client.providers.connectionAttempt(attempt?.uid ?? "", signal),
    refetchInterval: 2_000,
    retry: 3,
  });

  useProviderReturn(showToast, reload, providerLabel);
  useEffect(() => {
    if (connections.error) showToast(message(connections.error), "error");
  }, [connections.error, showToast]);
  useEffect(() => {
    if (attemptQuery.error) showToast(message(attemptQuery.error), "error");
  }, [attemptQuery.error, showToast]);
  useEffect(() => {
    const updated = attemptQuery.data;
    if (!updated) return;
    setAttempt(updated);
    if (updated.status === ProviderConnectionAttemptStatus.CONNECTED) {
      showToast(`${providerLabel(updated.providerId)}已连接`);
      void reload();
    }
  }, [attemptQuery.data, reload, showToast]);

  const connect = async (providerId: ProviderId) => {
    try {
      const created = await client.providers.beginConnection(providerId);
      if (created.challenge.case === "redirect") {
        window.location.assign(created.challenge.value.authorizationUrl);
        return;
      }
      setAttempt(created);
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const disconnect = async (connection: ProviderConnection) => {
    if (
      !(await confirm({
        title: "断开音乐账号",
        description: `断开${providerLabel(connection.providerId)}账号？本地歌曲不会被删除。`,
        confirmLabel: "断开账号",
        destructive: true,
      }))
    )
      return;
    try {
      await client.providers.disconnect(connection.providerId);
      await reload();
      showToast("账号已断开");
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  if (connections.isPending) return <LoadingIndicator label="正在读取音乐账号" />;
  return (
    <>
      <div className="overflow-hidden rounded-xl border bg-card/35">
        {(connections.data ?? []).map((connection) => (
          <ProviderAccountRow
            key={connection.providerId}
            connection={connection}
            onConnect={() => void connect(connection.providerId)}
            onDisconnect={() => void disconnect(connection)}
          />
        ))}
      </div>
      <p className="mt-4 text-sm text-muted-foreground">
        每个平台只连接一个账号。凭据在服务器加密保存，不会写入浏览器存储。
      </p>
      <ProviderLoginDialog attempt={attempt} onClose={() => setAttempt(undefined)} />
    </>
  );
}

function useProviderReturn(
  toast: (message: string, variant?: "default" | "error") => void,
  reload: () => Promise<void>,
  providerLabel: (providerId: ProviderId) => string,
) {
  useEffect(() => {
    const url = new URL(window.location.href);
    const providerId = url.searchParams.get("provider") ?? "";
    if (!providerId) return;
    const connected = url.searchParams.get("connection") === "connected";
    toast(
      connected ? `${providerLabel(providerId)}已连接` : `${providerLabel(providerId)}连接失败`,
      connected ? "default" : "error",
    );
    if (connected) void reload();
    url.searchParams.delete("provider");
    url.searchParams.delete("connection");
    window.history.replaceState(null, "", url);
  }, [providerLabel, reload, toast]);
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "操作未完成";
}
