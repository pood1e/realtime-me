import { useCallback, useEffect, useMemo, useState } from "react";
import { Download, ImageOff, Search, Sparkles, X } from "lucide-react";
import { WallpaperOrientation } from "@cloud-drive/contracts";
import type { Wallpaper } from "@cloud-drive/contracts";
import {
  Button,
  DialogRoot,
  DialogContent,
  EmptyState,
  Input,
  LoadingIndicator,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  WallpaperPublicClient,
} from "@cloud-drive/shared";

import { API_BASE } from "./config";

type OrientationFilter = "all" | "landscape" | "portrait" | "square";

export function App() {
  const client = useMemo(() => new WallpaperPublicClient(API_BASE), []);
  const [wallpapers, setWallpapers] = useState<Wallpaper[]>([]);
  const [query, setQuery] = useState("");
  const [orientation, setOrientation] = useState<OrientationFilter>("all");
  const [selected, setSelected] = useState<Wallpaper>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setWallpapers(
        await client.list({
          query,
          orientation: orientationValue(orientation),
        }),
      );
    } catch (loadError) {
      setError(errorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [client, orientation, query]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 180);
    return () => window.clearTimeout(timer);
  }, [load]);

  return (
    <main className="min-h-dvh bg-background text-foreground">
      <Hero
        query={query}
        orientation={orientation}
        onQueryChange={setQuery}
        onOrientationChange={setOrientation}
      />
      <section className="mx-auto w-full max-w-[112rem] px-4 py-8 sm:px-6 lg:px-8">
        {loading ? <LoadingIndicator label="正在加载壁纸" /> : null}
        {error ? (
          <EmptyState
            icon={<ImageOff className="size-6" />}
            title="壁纸暂时不可用"
            detail={error}
          />
        ) : null}
        {!loading && !error && wallpapers.length === 0 ? (
          <EmptyState title="没有匹配的壁纸" detail="换个关键词或方向试试。" />
        ) : null}
        {!loading && !error ? (
          <WallpaperGrid
            wallpapers={wallpapers}
            client={client}
            onSelect={setSelected}
          />
        ) : null}
      </section>
      <WallpaperPreview
        wallpaper={selected}
        client={client}
        onClose={() => setSelected(undefined)}
      />
    </main>
  );
}

function Hero({
  query,
  orientation,
  onQueryChange,
  onOrientationChange,
}: {
  query: string;
  orientation: OrientationFilter;
  onQueryChange: (value: string) => void;
  onOrientationChange: (value: OrientationFilter) => void;
}) {
  return (
    <header className="border-b bg-card/35">
      <div className="mx-auto max-w-[112rem] px-4 py-12 sm:px-6 lg:px-8 lg:py-20">
        <div className="max-w-2xl">
          <p className="flex items-center gap-2 text-xs font-semibold tracking-[0.2em] text-primary uppercase">
            <Sparkles className="size-4" /> Curated locally
          </p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight sm:text-6xl">
            每一块屏幕，
            <br />
            都值得一幅好画面。
          </h1>
          <p className="mt-5 max-w-xl text-base leading-7 text-muted-foreground">
            来自私人图库的精选壁纸，以适合屏幕的尺寸直接下载。
          </p>
        </div>
        <div className="mt-9 flex max-w-3xl flex-col gap-3 sm:flex-row">
          <div className="relative flex-1">
            <Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(event) => onQueryChange(event.target.value)}
              placeholder="搜索标题或标签"
              className="h-11 pl-9"
            />
          </div>
          <Select
            value={orientation}
            onValueChange={(value) =>
              onOrientationChange(value as OrientationFilter)
            }
          >
            <SelectTrigger className="h-11 w-full sm:w-44">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部方向</SelectItem>
              <SelectItem value="landscape">横向</SelectItem>
              <SelectItem value="portrait">竖向</SelectItem>
              <SelectItem value="square">方形</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
    </header>
  );
}

function WallpaperGrid({
  wallpapers,
  client,
  onSelect,
}: {
  wallpapers: readonly Wallpaper[];
  client: WallpaperPublicClient;
  onSelect: (wallpaper: Wallpaper) => void;
}) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
      {wallpapers.map((wallpaper) => {
        const preview = wallpaper.variants.at(0)?.url || wallpaper.originalUrl;
        return (
          <button
            key={wallpaper.uid}
            type="button"
            onClick={() => onSelect(wallpaper)}
            className="group relative aspect-[16/10] overflow-hidden rounded-2xl border bg-card text-left"
          >
            <img
              src={client.assetUrl(preview)}
              alt={wallpaper.title}
              loading="lazy"
              className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]"
            />
            <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/80 to-transparent p-5 pt-14">
              <h2 className="font-medium text-white">{wallpaper.title}</h2>
              <p className="mt-1 text-xs text-white/65">
                {wallpaper.width} × {wallpaper.height}
              </p>
            </div>
          </button>
        );
      })}
    </div>
  );
}

function WallpaperPreview({
  wallpaper,
  client,
  onClose,
}: {
  wallpaper?: Wallpaper;
  client: WallpaperPublicClient;
  onClose: () => void;
}) {
  const variant = wallpaper?.variants.at(-1);
  const downloadPath = variant?.url || wallpaper?.originalUrl || "";
  return (
    <DialogRoot
      open={Boolean(wallpaper)}
      onOpenChange={(open) => !open && onClose()}
    >
      <DialogContent
        showCloseButton={false}
        className="max-h-[calc(100dvh-1rem)] max-w-[calc(100vw-1rem)] overflow-hidden border-0 bg-black p-0 sm:max-w-6xl"
      >
        {wallpaper ? (
          <div className="relative">
            <img
              src={client.assetUrl(downloadPath)}
              alt={wallpaper.title}
              className="max-h-[84dvh] w-full object-contain"
            />
            <div className="flex items-center gap-3 border-t border-white/10 bg-black p-4 text-white">
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{wallpaper.title}</p>
                <p className="text-xs text-white/55">
                  {wallpaper.tags.join(" · ")}
                </p>
              </div>
              <Button asChild>
                <a href={`${client.assetUrl(downloadPath)}?download=1`}>
                  <Download />
                  下载
                </a>
              </Button>
              <Button variant="ghost" size="icon" onClick={onClose}>
                <X />
              </Button>
            </div>
          </div>
        ) : null}
      </DialogContent>
    </DialogRoot>
  );
}

function orientationValue(
  value: OrientationFilter,
): WallpaperOrientation | undefined {
  if (value === "landscape") return WallpaperOrientation.LANDSCAPE;
  if (value === "portrait") return WallpaperOrientation.PORTRAIT;
  if (value === "square") return WallpaperOrientation.SQUARE;
  return undefined;
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "暂时无法读取壁纸。";
}
