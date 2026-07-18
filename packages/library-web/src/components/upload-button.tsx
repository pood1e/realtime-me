import { Button } from "@realtime-me/web-ui/button";
import { LoaderCircle, Upload } from "lucide-react";
import { useRef, useState } from "react";

export function UploadButton({
  accept,
  multiple = true,
  label = "上传",
  onFiles,
}: {
  accept?: string;
  multiple?: boolean;
  label?: string;
  onFiles: (files: File[]) => Promise<void>;
}) {
  const input = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const select = async (files: FileList | null) => {
    if (!files?.length) return;
    setBusy(true);
    try {
      await onFiles(Array.from(files));
    } finally {
      setBusy(false);
      if (input.current) input.current.value = "";
    }
  };
  return (
    <>
      <input
        ref={input}
        hidden
        type="file"
        accept={accept}
        multiple={multiple}
        onChange={(event) => void select(event.target.files)}
      />
      <Button disabled={busy} onClick={() => input.current?.click()}>
        {busy ? <LoaderCircle className="size-4 animate-spin" /> : <Upload className="size-4" />}
        {busy ? "上传中" : label}
      </Button>
    </>
  );
}
