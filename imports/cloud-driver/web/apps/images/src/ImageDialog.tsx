import { useState } from "react";
import { Check, Copy, Download, ExternalLink, Palette } from "lucide-react";
import type { Image } from "@cloud-drive/contracts";
import {
  Button,
  Dialog,
  ImagesClient,
  Input,
  WallpaperAdminClient,
  useToast,
} from "@cloud-drive/shared";
import { API_BASE } from "./config";

export function ImageDialog({
  image,
  client,
  onClose,
  onChanged,
}: {
  image: Image;
  client: ImagesClient;
  onClose: () => void;
  onChanged: () => Promise<void>;
}) {
  const wallpapers = new WallpaperAdminClient(API_BASE);
  const { showToast } = useToast();
  const [link, setLink] = useState("");
  const [publishing, setPublishing] = useState(false);
  const createLink = async () => {
    try {
      const created = await client.createLink(image.uid);
      setLink(created.publicUrl);
      await navigator.clipboard.writeText(created.publicUrl);
      showToast("匿名原图链接已复制");
    } catch (error) {
      showToast(message(error), "error");
    }
  };
  const publish = async () => {
    const title = window.prompt("壁纸标题", image.displayName);
    if (!title?.trim()) return;
    const rawTags = window.prompt("标签（用逗号分隔）", "");
    setPublishing(true);
    try {
      await wallpapers.publish(
        image.uid,
        title.trim(),
        (rawTags || "")
          .split(/[,，]/)
          .map((tag) => tag.trim())
          .filter(Boolean),
      );
      showToast("已发布到壁纸站");
    } catch (error) {
      showToast(message(error), "error");
    } finally {
      setPublishing(false);
    }
  };
  const copyOriginal = async () => {
    await navigator.clipboard.writeText(client.originalUrl(image));
    showToast("私有原图地址已复制");
  };
  return (
    <Dialog
      open
      title={image.displayName}
      description={`${image.width} × ${image.height}`}
      size="preview"
      onClose={onClose}
    >
      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="grid max-h-[72dvh] place-items-center overflow-auto rounded-xl bg-muted/35 p-2">
          <img
            src={client.originalUrl(image)}
            alt={image.displayName}
            className="max-h-[68dvh] max-w-full object-contain"
          />
        </div>
        <aside className="space-y-3">
          <Button className="w-full" onClick={() => void createLink()}>
            <ExternalLink />
            生成匿名原图链接
          </Button>
          {link ? (
            <div className="space-y-2">
              <Input readOnly value={link} />
              <Button
                variant="outline"
                className="w-full"
                onClick={() =>
                  void navigator.clipboard
                    .writeText(link)
                    .then(() => showToast("链接已复制"))
                }
              >
                <Check />
                再次复制
              </Button>
            </div>
          ) : null}
          <Button
            variant="outline"
            className="w-full"
            disabled={publishing}
            onClick={() => void publish()}
          >
            <Palette />
            {publishing ? "发布中" : "发布为壁纸"}
          </Button>
          <Button
            variant="ghost"
            className="w-full"
            onClick={() => void copyOriginal()}
          >
            <Copy />
            复制私有地址
          </Button>
          <Button variant="ghost" className="w-full" asChild>
            <a href={`${client.originalUrl(image)}?download=1`}>
              <Download />
              下载原图
            </a>
          </Button>
        </aside>
      </div>
    </Dialog>
  );
}
function message(error: unknown) {
  return error instanceof Error ? error.message : "操作未完成";
}
