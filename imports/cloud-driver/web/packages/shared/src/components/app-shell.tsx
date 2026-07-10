import type { PropsWithChildren, ReactNode } from "react";
import {
  BookOpen,
  HardDrive,
  Image,
  LogOut,
  Menu,
  Music,
  Palette,
  Share2,
} from "lucide-react";

import { SessionClient } from "../api";
import type { AppLinks, PrivateAppId } from "../configuration";
import { cn } from "../lib/utils";
import { Button } from "./ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./ui/sheet";

const privateApps = [
  { id: "drive" as const, label: "云盘", icon: HardDrive },
  { id: "books" as const, label: "书架", icon: BookOpen },
  { id: "music" as const, label: "音乐盒", icon: Music },
  { id: "images" as const, label: "图床", icon: Image },
];

function AppNavigation({
  current,
  links,
}: {
  current: PrivateAppId;
  links: AppLinks;
}) {
  return (
    <nav className="space-y-1" aria-label="应用">
      {privateApps.map(({ id, label, icon: Icon }) => {
        const href = links[id];
        const active = current === id;
        return href ? (
          <a
            key={id}
            href={href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "flex h-10 items-center gap-3 rounded-lg px-3 text-sm transition-colors",
              active
                ? "bg-primary/12 text-primary"
                : "text-muted-foreground hover:bg-accent hover:text-foreground",
            )}
          >
            <Icon className="size-4" />
            {label}
          </a>
        ) : null;
      })}
      <div className="my-3 border-t" />
      {links.wallpapers ? (
        <a
          href={links.wallpapers}
          className="flex h-10 items-center gap-3 rounded-lg px-3 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        >
          <Palette className="size-4" />
          壁纸站
        </a>
      ) : null}
      {links.share ? (
        <a
          href={links.share}
          className="flex h-10 items-center gap-3 rounded-lg px-3 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        >
          <Share2 className="size-4" />
          分享页
        </a>
      ) : null}
    </nav>
  );
}

export function PrivateAppShell({
  app,
  title,
  subtitle,
  apiBase,
  links,
  actions,
  children,
}: PropsWithChildren<{
  app: PrivateAppId;
  title: string;
  subtitle?: string;
  apiBase: string;
  links: AppLinks;
  actions?: ReactNode;
}>) {
  const logout = async () => {
    await new SessionClient(apiBase).logout();
    window.location.reload();
  };
  const navigation = <AppNavigation current={app} links={links} />;
  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[15rem_minmax(0,1fr)]">
      <aside className="fixed inset-y-0 left-0 hidden w-60 border-r bg-card/45 p-5 lg:block">
        <p className="mb-7 px-3 text-xs font-semibold tracking-[0.2em] text-primary uppercase">
          Local Library
        </p>
        {navigation}
        <Button
          variant="ghost"
          className="absolute right-5 bottom-5 left-5 justify-start text-muted-foreground"
          onClick={() => void logout()}
        >
          <LogOut className="size-4" />
          退出登录
        </Button>
      </aside>
      <main className="min-w-0 lg:col-start-2">
        <header className="sticky top-0 z-30 flex min-h-16 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur sm:px-6 lg:px-8">
          <Sheet>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" className="lg:hidden">
                <Menu className="size-5" />
                <span className="sr-only">打开导航</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-72 p-5">
              <SheetHeader className="px-3">
                <SheetTitle className="text-left text-sm tracking-[0.18em] text-primary uppercase">
                  Local Library
                </SheetTitle>
              </SheetHeader>
              <div className="mt-6">{navigation}</div>
              <Button
                variant="ghost"
                className="mt-8 w-full justify-start text-muted-foreground"
                onClick={() => void logout()}
              >
                <LogOut className="size-4" />
                退出登录
              </Button>
            </SheetContent>
          </Sheet>
          <div className="min-w-0 flex-1">
            <h1 className="truncate text-base font-semibold">{title}</h1>
            {subtitle ? (
              <p className="truncate text-xs text-muted-foreground">
                {subtitle}
              </p>
            ) : null}
          </div>
          <div className="flex shrink-0 items-center gap-2">{actions}</div>
        </header>
        <div className="mx-auto w-full max-w-[112rem] p-4 sm:p-6 lg:p-8">
          {children}
        </div>
      </main>
    </div>
  );
}
