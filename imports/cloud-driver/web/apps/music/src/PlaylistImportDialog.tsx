import { useState, type FormEvent } from "react";
import { ListPlus } from "lucide-react";
import { MusicProvider } from "@cloud-drive/contracts";
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
} from "@cloud-drive/shared";

export function PlaylistImportDialog({
  importing,
  onImport,
}: {
  importing: boolean;
  onImport: (provider: MusicProvider, source: string) => Promise<boolean>;
}) {
  const [open, setOpen] = useState(false);
  const [provider, setProvider] = useState(MusicProvider.QQ_MUSIC);
  const [source, setSource] = useState("");
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!source.trim()) return;
    if (!(await onImport(provider, source.trim()))) return;
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
            <Select
              value={String(provider)}
              onValueChange={(value) => setProvider(Number(value))}
            >
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={String(MusicProvider.QQ_MUSIC)}>
                  QQ 音乐
                </SelectItem>
                <SelectItem value={String(MusicProvider.NETEASE_CLOUD_MUSIC)}>
                  网易云音乐
                </SelectItem>
                <SelectItem value={String(MusicProvider.SPOTIFY)}>
                  Spotify
                </SelectItem>
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
            <Button type="submit" disabled={importing || !source.trim()}>
              {importing ? "正在导入" : "导入"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </DialogRoot>
  );
}
