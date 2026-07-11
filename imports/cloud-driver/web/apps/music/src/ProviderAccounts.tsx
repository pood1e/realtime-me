import { useCallback, useEffect, useState } from "react";
import {
  MusicProvider,
  ProviderConnectionAttemptStatus,
  type ProviderConnection,
  type ProviderConnectionAttempt,
} from "@cloud-drive/contracts";
import { LoadingIndicator, MusicClient, useToast } from "@cloud-drive/shared";
import { providerLabel } from "./music-model";
import { ProviderAccountRow } from "./ProviderAccountRow";
import { ProviderLoginDialog } from "./ProviderLoginDialog";
import { terminalConnectionAttempt } from "./provider-account-model";

export function ProviderAccounts({ client }: { client: MusicClient }) {
  const { showToast } = useToast();
  const [connections, setConnections] = useState<ProviderConnection[]>([]);
  const [attempt, setAttempt] = useState<ProviderConnectionAttempt>();
  const [loading, setLoading] = useState(true);
  const load = useCallback(async () => {
    setLoading(true);
    try {
      setConnections(await client.providerConnections());
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  }, [client, showToast]);
  useEffect(() => void load(), [load]);
  useSpotifyReturn(showToast, load);
  useEffect(() => {
    if (!attempt || terminalConnectionAttempt(attempt.status)) return;
    const timer = window.setTimeout(() => {
      void client
        .providerConnectionAttempt(attempt.uid)
        .then((updated) => {
          setAttempt(updated);
          if (updated.status === ProviderConnectionAttemptStatus.CONNECTED) {
            showToast(`${providerLabel(updated.provider)}已连接`);
            void load();
          }
        })
        .catch((error: unknown) => showToast(message(error), "error"));
    }, 2_000);
    return () => window.clearTimeout(timer);
  }, [attempt, client, load, showToast]);
  const connect = async (provider: MusicProvider) => {
    try {
      const created = await client.beginProviderConnection(provider);
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
    if (!window.confirm(`断开${providerLabel(connection.provider)}账号？`))
      return;
    try {
      await client.disconnectProvider(connection.provider);
      await load();
      showToast("账号已断开");
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  if (loading) return <LoadingIndicator label="正在读取音乐账号" />;
  return (
    <>
      <div className="overflow-hidden rounded-xl border bg-card/35">
        {connections.map((connection) => (
          <ProviderAccountRow
            key={connection.provider}
            connection={connection}
            onConnect={() => void connect(connection.provider)}
            onDisconnect={() => void disconnect(connection)}
          />
        ))}
      </div>
      <p className="mt-4 text-sm text-muted-foreground">
        每个平台只连接一个账号。凭据在服务器加密保存，不会写入浏览器存储。
      </p>
      <ProviderLoginDialog
        attempt={attempt}
        onClose={() => setAttempt(undefined)}
      />
    </>
  );
}

function useSpotifyReturn(
  toast: (message: string, variant?: "default" | "error") => void,
  reload: () => Promise<void>,
) {
  useEffect(() => {
    const url = new URL(window.location.href);
    if (url.searchParams.get("provider") !== "spotify") return;
    const connected = url.searchParams.get("connection") === "connected";
    toast(
      connected ? "Spotify 已连接" : "Spotify 连接失败",
      connected ? "default" : "error",
    );
    if (connected) void reload();
    url.searchParams.delete("provider");
    url.searchParams.delete("connection");
    window.history.replaceState(null, "", url);
  }, [reload, toast]);
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "操作未完成";
}
