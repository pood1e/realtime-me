import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ImageIcon,
  Images,
  MoreHorizontal,
  Plus,
  Search,
  Trash2,
} from "lucide-react";
import type { Image, ImageAlbum } from "@cloud-drive/contracts";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  EmptyState,
  ImagesClient,
  Input,
  LoadingIndicator,
  PrivateAppShell,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  UploadButton,
  UploadClient,
  formatBytes,
  useToast,
} from "@cloud-drive/shared";
import { ImageDialog } from "./ImageDialog";
import { API_BASE, APP_LINKS } from "./config";

export function ImagesPage() {
  const client = useMemo(() => new ImagesClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const [images, setImages] = useState<Image[]>([]);
  const [albums, setAlbums] = useState<ImageAlbum[]>([]);
  const [albumUid, setAlbumUid] = useState("all");
  const [query, setQuery] = useState("");
  const [trash, setTrash] = useState(false);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Image>();
  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [nextImages, nextAlbums] = await Promise.all([
        client.list({
          query,
          albumUid: albumUid === "all" ? undefined : albumUid,
          trashed: trash,
        }),
        client.albums(),
      ]);
      setImages(nextImages);
      setAlbums(nextAlbums);
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setLoading(false);
    }
  }, [albumUid, client, query, showToast, trash]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timer);
  }, [load]);
  const upload = async (files: File[]) => {
    for (const file of files)
      try {
        const uid = await uploader.upload(file);
        await client.importUpload(uid, albumUid === "all" ? "" : albumUid);
        showToast(`${file.name} 已上传`);
      } catch (error) {
        showToast(`${file.name}: ${message(error)}`, "error");
      }
    await load();
  };
  const createAlbum = async () => {
    const name = window.prompt("相册名称");
    if (!name?.trim()) return;
    try {
      const album = await client.createAlbum(name.trim());
      setAlbumUid(album.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (image: Image) => {
    try {
      if (trash) {
        if (!window.confirm("永久删除此图片？")) return;
        await client.purge(image.uid);
      } else await client.trash(image.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const restore = async (image: Image) => {
    await client.restore(image.uid);
    await load();
  };
  return (
    <PrivateAppShell
      app="images"
      title="图床"
      subtitle="图片管理、匿名直链与壁纸发布"
      apiBase={API_BASE}
      links={APP_LINKS}
      actions={
        trash ? (
          <Button
            variant="destructive"
            onClick={() => void emptyTrash(client, load, showToast)}
          >
            清空
          </Button>
        ) : (
          <UploadButton
            accept="image/jpeg,image/png,image/webp,image/gif,image/svg+xml"
            onFiles={upload}
            label="上传图片"
          />
        )
      }
    >
      <div className="mb-7 flex flex-col gap-3 xl:flex-row xl:items-center">
        <div className="relative min-w-0 flex-1">
          <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索图片"
            className="pl-9"
          />
        </div>
        <Select value={albumUid} onValueChange={setAlbumUid}>
          <SelectTrigger className="w-full xl:w-52">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部图片</SelectItem>
            {albums.map((album) => (
              <SelectItem key={album.uid} value={album.uid}>
                {album.displayName} · {album.imageCount}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button variant="outline" onClick={() => void createAlbum()}>
          <Plus />
          新建相册
        </Button>
        <Button
          variant={trash ? "secondary" : "ghost"}
          onClick={() => setTrash((value) => !value)}
        >
          <Trash2 />
          {trash ? "返回图库" : "回收站"}
        </Button>
      </div>
      {loading ? (
        <LoadingIndicator label="正在读取图片" />
      ) : images.length ? (
        <div className="columns-2 gap-3 sm:columns-3 lg:columns-4 2xl:columns-6">
          {images.map((image) => (
            <ImageCard
              key={image.uid}
              image={image}
              client={client}
              trashed={trash}
              onOpen={() => setSelected(image)}
              onRemove={() => void remove(image)}
              onRestore={() => void restore(image)}
            />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={<Images className="size-6" />}
          title="图库还是空的"
          detail="上传 JPEG、PNG、WebP、GIF 或 SVG，系统会生成预览并保留原图。"
        />
      )}
      {selected ? (
        <ImageDialog
          image={selected}
          client={client}
          onClose={() => setSelected(undefined)}
          onChanged={load}
        />
      ) : null}
    </PrivateAppShell>
  );
}

function ImageCard({
  image,
  client,
  trashed,
  onOpen,
  onRemove,
  onRestore,
}: {
  image: Image;
  client: ImagesClient;
  trashed: boolean;
  onOpen: () => void;
  onRemove: () => void;
  onRestore: () => void;
}) {
  return (
    <article className="group relative mb-3 break-inside-avoid overflow-hidden rounded-xl border bg-card">
      <button
        type="button"
        onClick={onOpen}
        disabled={trashed}
        className="block w-full bg-muted"
      >
        {image.previewUrl ? (
          <img
            src={client.previewUrl(image)}
            alt={image.displayName}
            loading="lazy"
            className="h-auto w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
          />
        ) : (
          <div className="grid aspect-square place-items-center">
            <ImageIcon className="size-8 text-muted-foreground" />
          </div>
        )}
      </button>
      <div className="flex items-center gap-2 p-3">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{image.displayName}</p>
          <p className="mt-0.5 text-[11px] text-muted-foreground">
            {image.width} × {image.height} · {formatBytes(image.sizeBytes)}
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon-xs">
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {trashed ? (
              <DropdownMenuItem onSelect={onRestore}>恢复</DropdownMenuItem>
            ) : (
              <DropdownMenuItem onSelect={onOpen}>打开与分享</DropdownMenuItem>
            )}
            <DropdownMenuItem variant="destructive" onSelect={onRemove}>
              {trashed ? "永久删除" : "移入回收站"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </article>
  );
}
function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成";
}
async function emptyTrash(
  client: ImagesClient,
  reload: () => Promise<void>,
  toast: (message: string, variant?: "default" | "error") => void,
) {
  if (!window.confirm("永久删除图片回收站？")) return;
  try {
    await client.emptyTrash();
    await reload();
    toast("回收站已清空");
  } catch (error) {
    toast(message(error), "error");
  }
}
