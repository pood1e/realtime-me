import { useEffect, useState } from "react";
import {
  MusicProvider,
  type ProviderConnectionAttempt,
} from "@cloud-drive/contracts";
import { Dialog } from "@cloud-drive/shared";
import { QRCodeSVG } from "qrcode.react";
import { providerLabel } from "./music-model";
import {
  connectionAttemptStatus,
  terminalConnectionAttempt,
} from "./provider-account-model";

export function ProviderLoginDialog({
  attempt,
  onClose,
}: {
  attempt?: ProviderConnectionAttempt;
  onClose: () => void;
}) {
  const qr = attempt?.challenge.case === "qr" ? attempt.challenge.value : null;
  const imageURL = useQRImage(qr?.image, qr?.contentType);
  return (
    <Dialog
      open={Boolean(attempt)}
      title={`连接${providerLabel(attempt?.provider ?? MusicProvider.UNSPECIFIED)}`}
      description={connectionAttemptStatus(attempt?.status)}
      onClose={onClose}
    >
      {qr ? (
        <div className="mx-auto rounded-xl bg-white p-4">
          {imageURL ? (
            <img src={imageURL} alt="登录二维码" className="size-56" />
          ) : qr.payload ? (
            <QRCodeSVG value={qr.payload} size={224} level="M" />
          ) : null}
        </div>
      ) : null}
      {attempt && !terminalConnectionAttempt(attempt.status) ? (
        <p className="text-center text-xs text-muted-foreground">
          页面会自动检查扫码结果
        </p>
      ) : null}
    </Dialog>
  );
}

function useQRImage(
  image: Uint8Array | undefined,
  contentType: string | undefined,
) {
  const [url, setURL] = useState("");
  useEffect(() => {
    if (!image?.length) {
      setURL("");
      return;
    }
    const nextURL = URL.createObjectURL(
      new Blob([new Uint8Array(image).buffer], {
        type: contentType || "image/png",
      }),
    );
    setURL(nextURL);
    return () => URL.revokeObjectURL(nextURL);
  }, [contentType, image]);
  return url;
}
