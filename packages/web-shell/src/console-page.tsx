import type { PropsWithChildren, ReactNode } from "react";

// ConsolePage gives every bounded-context screen one consistent content header.
export function ConsolePage({
  title,
  subtitle,
  actions,
  children,
}: PropsWithChildren<{
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}>) {
  return (
    <main className="min-w-0">
      <header className="sticky top-0 z-30 flex min-h-16 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur sm:px-6 lg:px-8">
        <div className="min-w-0 flex-1 pl-11 lg:pl-0">
          <h1 className="truncate text-base font-semibold">{title}</h1>
          {subtitle ? <p className="truncate text-xs text-muted-foreground">{subtitle}</p> : null}
        </div>
        {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
      </header>
      <div className="mx-auto w-full max-w-[112rem] p-4 sm:p-6 lg:p-8">{children}</div>
    </main>
  );
}
