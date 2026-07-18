import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type FormEvent,
  type PropsWithChildren,
  type ReactNode,
} from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";
import { Button } from "./ui/button";
import {
  Dialog as DialogRoot,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Input } from "./ui/input";

export type ConfirmDialogOptions = Readonly<{
  title: string;
  description: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
}>;

export type PromptDialogOptions = Readonly<{
  title: string;
  description?: ReactNode;
  label: string;
  defaultValue?: string;
  placeholder?: string;
  submitLabel?: string;
  required?: boolean;
}>;

type ConfirmRequest = Readonly<{
  id: number;
  kind: "confirm";
  options: ConfirmDialogOptions;
  resolve: (value: boolean) => void;
}>;

type PromptRequest = Readonly<{
  id: number;
  kind: "prompt";
  options: PromptDialogOptions;
  resolve: (value: string | undefined) => void;
}>;

type DialogRequest = ConfirmRequest | PromptRequest;

type DialogContextValue = Readonly<{
  confirm: (options: ConfirmDialogOptions) => Promise<boolean>;
  prompt: (options: PromptDialogOptions) => Promise<string | undefined>;
}>;

const DialogContext = createContext<DialogContextValue | undefined>(undefined);

export function DialogProvider({ children }: PropsWithChildren) {
  const sequence = useRef(0);
  const queue = useRef<DialogRequest[]>([]);
  const [active, setActive] = useState<DialogRequest>();

  const enqueue = useCallback((request: DialogRequest) => {
    setActive((current) => {
      if (!current) return request;
      queue.current.push(request);
      return current;
    });
  }, []);
  const finish = useCallback((request: DialogRequest) => {
    setActive((current) => {
      if (current?.id !== request.id) return current;
      return queue.current.shift();
    });
  }, []);
  const confirm = useCallback(
    (options: ConfirmDialogOptions) =>
      new Promise<boolean>((resolve) => {
        enqueue({
          id: ++sequence.current,
          kind: "confirm",
          options,
          resolve,
        });
      }),
    [enqueue],
  );
  const prompt = useCallback(
    (options: PromptDialogOptions) =>
      new Promise<string | undefined>((resolve) => {
        enqueue({
          id: ++sequence.current,
          kind: "prompt",
          options,
          resolve,
        });
      }),
    [enqueue],
  );
  const value = useMemo(() => ({ confirm, prompt }), [confirm, prompt]);

  return (
    <DialogContext.Provider value={value}>
      {children}
      {active?.kind === "confirm" ? (
        <ConfirmDialog
          key={active.id}
          request={active}
          onFinish={() => finish(active)}
        />
      ) : null}
      {active?.kind === "prompt" ? (
        <PromptDialog
          key={active.id}
          request={active}
          onFinish={() => finish(active)}
        />
      ) : null}
    </DialogContext.Provider>
  );
}

export function useDialog(): DialogContextValue {
  const context = useContext(DialogContext);
  if (!context) throw new Error("useDialog must be used inside DialogProvider");
  return context;
}

function ConfirmDialog({
  request,
  onFinish,
}: {
  request: ConfirmRequest;
  onFinish: () => void;
}) {
  const settled = useRef(false);
  const settle = (value: boolean) => {
    if (settled.current) return;
    settled.current = true;
    request.resolve(value);
    onFinish();
  };
  return (
    <AlertDialog open onOpenChange={(open) => !open && settle(false)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{request.options.title}</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div>{request.options.description}</div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => settle(false)}>
            {request.options.cancelLabel ?? "取消"}
          </AlertDialogCancel>
          <AlertDialogAction
            variant={request.options.destructive ? "destructive" : "default"}
            onClick={() => settle(true)}
          >
            {request.options.confirmLabel ?? "确认"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function PromptDialog({
  request,
  onFinish,
}: {
  request: PromptRequest;
  onFinish: () => void;
}) {
  const [value, setValue] = useState(request.options.defaultValue ?? "");
  const settled = useRef(false);
  const settle = (result: string | undefined) => {
    if (settled.current) return;
    settled.current = true;
    request.resolve(result);
    onFinish();
  };
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const normalized = value.trim();
    if (normalized || request.options.required === false) settle(normalized);
  };
  return (
    <DialogRoot open onOpenChange={(open) => !open && settle(undefined)}>
      <DialogContent>
        <form className="space-y-4" onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>{request.options.title}</DialogTitle>
            {request.options.description ? (
              <DialogDescription asChild>
                <div>{request.options.description}</div>
              </DialogDescription>
            ) : null}
          </DialogHeader>
          <label className="grid gap-2 text-sm font-medium">
            {request.options.label}
            <Input
              autoFocus
              value={value}
              {...(request.options.placeholder
                ? { placeholder: request.options.placeholder }
                : {})}
              onChange={(event) => setValue(event.target.value)}
            />
          </label>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => settle(undefined)}
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={request.options.required !== false && !value.trim()}
            >
              {request.options.submitLabel ?? "确认"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </DialogRoot>
  );
}
