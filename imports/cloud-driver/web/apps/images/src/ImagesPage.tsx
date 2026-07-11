import { useDeferredValue, useEffect, useMemo, useState } from "react";
import { Images, Plus, Search, Trash2 } from "lucide-react";
import { ProcessingStatus, type Image } from "@cloud-drive/contracts";
import {
  Button,
  EmptyState,
  ImagesClient,
  InfiniteScrollSentinel,
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
  useCursorQuery,
  useDialog,
  useQuery,
  useToast,
} from "@cloud-drive/shared";
import { ImageDialog } from "./ImageDialog";
import { ImageCard } from "./ImageCard";
import { API_BASE, APP_LINKS } from "./config";

export function ImagesPage() {
  const client = useMemo(() => new ImagesClient(API_BASE), []);
  const uploader = useMemo(() => new UploadClient(API_BASE), []);
  const { showToast } = useToast();
  const { confirm, prompt } = useDialog();
  const [albumUid, setAlbumUid] = useState("all");
  const [query, setQuery] = useState("");
  const [trash, setTrash] = useState(false);
  const [selected, setSelected] = useState<Image>();
  const deferredQuery = useDeferredValue(query.trim());
  const catalog = useCursorQuery<Image>({
    queryKey: ["images", deferredQuery, albumUid, trash],
    pollInterval: 2_500,
    shouldPoll: (images) =>
      images.some(
        (image) => image.processingStatus === ProcessingStatus.PENDING,
      ),
    loadPage: async (pageToken, signal) => {
      const page = await client.listPage(
        {
          query: deferredQuery,
          ...(albumUid === "all" ? {} : { albumUid }),
          trashed: trash,
          pageToken,
        },
        signal,
      );
      return { items: page.images, nextPageToken: page.nextPageToken };
    },
  });
  const albumCatalog = useQuery({
    queryKey: ["image-albums"],
    queryFn: ({ signal }) => client.albums(signal),
  });
  const load = async () => {
    await Promise.all([catalog.refetch(), albumCatalog.refetch()]);
  };
  useEffect(() => {
    if (catalog.error) showToast(message(catalog.error), "error");
  }, [catalog.error, showToast]);
  useEffect(() => {
    if (albumCatalog.error) showToast(message(albumCatalog.error), "error");
  }, [albumCatalog.error, showToast]);
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
    const name = await prompt({ title: "新建相册", label: "相册名称" });
    if (!name) return;
    try {
      const album = await client.createAlbum(name);
      setAlbumUid(album.uid);
      await load();
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const remove = async (image: Image) => {
    try {
      if (trash) {
        if (
          !(await confirm({
            title: "永久删除图片",
            description: `“${image.displayName}”将无法恢复。`,
            confirmLabel: "永久删除",
            destructive: true,
          }))
        )
          return;
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
  const emptyTrash = async () => {
    if (
      !(await confirm({
        title: "清空图片回收站",
        description: "回收站中的全部图片将被永久删除，此操作无法撤销。",
        confirmLabel: "永久删除全部图片",
        destructive: true,
      }))
    )
      return;
    try {
      await client.emptyTrash();
      await load();
      showToast("回收站已清空");
    } catch (error) {
      showToast(message(error), "error");
    }
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
          <Button variant="destructive" onClick={() => void emptyTrash()}>
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
            {(albumCatalog.data ?? []).map((album) => (
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
      {catalog.isPending ? (
        <LoadingIndicator label="正在读取图片" />
      ) : catalog.items.length ? (
        <>
          <div className="columns-2 gap-3 sm:columns-3 lg:columns-4 2xl:columns-6">
            {catalog.items.map((image) => (
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
          <InfiniteScrollSentinel
            hasMore={catalog.hasNextPage}
            loading={catalog.isFetchingNextPage}
            failed={catalog.isFetchNextPageError}
            loadingLabel="继续加载图片"
            completeLabel="已加载全部图片"
            onLoadMore={() => void catalog.fetchNextPage()}
          />
        </>
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
        />
      ) : null}
    </PrivateAppShell>
  );
}

function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成";
}
