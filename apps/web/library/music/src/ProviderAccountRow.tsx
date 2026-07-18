import { Link2, RefreshCw, Unlink } from "lucide-react";
import {
  ProviderConnectionStatus,
  type ProviderConnection,
} from "@realtime-me/library-contracts";
import { Badge, Button } from "@realtime-me/library-web";
import { useProviderLabel } from "./provider-catalog";
import { providerConnectionDetail } from "./provider-account-model";

export function ProviderAccountRow({
  connection,
  onConnect,
  onDisconnect,
}: {
  connection: ProviderConnection;
  onConnect: () => void;
  onDisconnect: () => void;
}) {
  const providerLabel = useProviderLabel();
  const connected = connection.status === ProviderConnectionStatus.CONNECTED;
  const reconnect =
    connection.status === ProviderConnectionStatus.RECONNECT_REQUIRED;
  return (
    <div className="flex flex-col gap-4 border-b px-5 py-5 last:border-b-0 sm:flex-row sm:items-center">
      <div className="flex min-w-0 flex-1 items-center gap-4">
        {connection.avatarUrl ? (
          <img
            src={connection.avatarUrl}
            alt=""
            className="size-11 rounded-full object-cover"
          />
        ) : (
          <div className="grid size-11 place-items-center rounded-full bg-muted text-sm font-semibold">
            {providerLabel(connection.providerId).slice(0, 1)}
          </div>
        )}
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-medium">
              {providerLabel(connection.providerId)}
            </p>
            <ConnectionBadge status={connection.status} />
          </div>
          <p className="truncate text-sm text-muted-foreground">
            {connection.displayName ||
              providerConnectionDetail(connection.status)}
          </p>
          {connection.membership ? (
            <p className="mt-1 text-xs text-muted-foreground">
              {connection.membership}
            </p>
          ) : null}
        </div>
      </div>
      {connected ? (
        <Button variant="outline" onClick={onDisconnect}>
          <Unlink />
          断开
        </Button>
      ) : (
        <Button
          onClick={onConnect}
          disabled={
            connection.status === ProviderConnectionStatus.NOT_CONFIGURED
          }
        >
          {reconnect ? <RefreshCw /> : <Link2 />}
          {reconnect ? "重新连接" : "连接账号"}
        </Button>
      )}
    </div>
  );
}

function ConnectionBadge({ status }: { status: ProviderConnectionStatus }) {
  if (status === ProviderConnectionStatus.CONNECTED)
    return <Badge variant="secondary">已连接</Badge>;
  if (status === ProviderConnectionStatus.RECONNECT_REQUIRED)
    return <Badge variant="destructive">需要重连</Badge>;
  if (status === ProviderConnectionStatus.NOT_CONFIGURED)
    return <Badge variant="outline">未配置</Badge>;
  return <Badge variant="outline">未连接</Badge>;
}
