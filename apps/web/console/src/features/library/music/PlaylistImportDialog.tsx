import type { ProviderDescriptor } from "@realtime-me/library-contracts";
import type { ProviderId } from "@realtime-me/library-web";
import {
  Button,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogRoot,
  DialogTitle,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@realtime-me/web-ui";
import { ListPlus } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useProviderLabel } from "./provider-catalog";

export function PlaylistImportDialog({
  importing,
  providers,
  onImport,
}: {
  importing: boolean;
  providers: ProviderDescriptor[];
  onImport: (providerId: ProviderId, source: string) => Promise<boolean>;
}) {
  const providerLabel = useProviderLabel();
  const [open, setOpen] = useState(false);
  const [providerId, setProviderId] = useState("");
  const [source, setSource] = useState("");
  useEffect(() => {
    if (!providers.some((provider) => provider.id === providerId)) {
      setProviderId(providers[0]?.id ?? "");
    }
  }, [providerId, providers]);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!source.trim()) return;
    if (!providerId || !(await onImport(providerId, source.trim()))) return;
    setSource("");
    setOpen(false);
  };
  return (
    <DialogRoot open={open} onOpenChange={setOpen}>
      <Button onClick={() => setOpen(true)}>
        <ListPlus />
        导入歌单
      </Button>
      <DialogContent>
        <form onSubmit={(event) => void submit(event)}>
          <DialogHeader>
            <DialogTitle>导入在线歌单</DialogTitle>
            <DialogDescription>
              粘贴歌单链接或歌单 ID，导入后可统一播放和存入本地。
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-5 sm:grid-cols-[10rem_minmax(0,1fr)]">
            <Select value={providerId} onValueChange={setProviderId}>
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {providers.map((provider) => (
                  <SelectItem key={provider.id} value={provider.id}>
                    {provider.displayName || providerLabel(provider.id)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Input
              autoFocus
              value={source}
              onChange={(event) => setSource(event.target.value)}
              placeholder="歌单链接或 ID"
            />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={importing || !providerId || !source.trim()}>
              {importing ? "正在导入" : "导入"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </DialogRoot>
  );
}
