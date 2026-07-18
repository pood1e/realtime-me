import { Permission } from "@realtime-me/auth-contracts";
import { Button } from "@realtime-me/web-ui/button";
import { cn } from "@realtime-me/web-ui/cn";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@realtime-me/web-ui/sheet";
import { BookOpen, Bot, HardDrive, Image, LogOut, Menu, Music, Server } from "lucide-react";
import type { ComponentType } from "react";
import { NavLink, Outlet } from "react-router-dom";
import { ThemeToggle } from "./theme";

type NavigationItem = Readonly<{
  label: string;
  to: string;
  icon: ComponentType<{ className?: string }>;
  permission: Permission;
}>;

const navigationGroups: ReadonlyArray<
  Readonly<{ label: string; items: readonly NavigationItem[] }>
> = [
  {
    label: "Operations",
    items: [
      {
        label: "Status",
        to: "/status",
        icon: Server,
        permission: Permission.STATUS_INTERNAL_READ,
      },
    ],
  },
  {
    label: "Library",
    items: [
      {
        label: "Drive",
        to: "/library/drive",
        icon: HardDrive,
        permission: Permission.LIBRARY_MANAGE,
      },
      {
        label: "Books",
        to: "/library/books",
        icon: BookOpen,
        permission: Permission.LIBRARY_MANAGE,
      },
      {
        label: "Music",
        to: "/library/music",
        icon: Music,
        permission: Permission.LIBRARY_MANAGE,
      },
      {
        label: "Images",
        to: "/library/images",
        icon: Image,
        permission: Permission.LIBRARY_MANAGE,
      },
    ],
  },
  {
    label: "Manager",
    items: [
      {
        label: "Agents",
        to: "/manager",
        icon: Bot,
        permission: Permission.MANAGER_CONTROL,
      },
    ],
  },
];

export function ConsoleShell({
  displayName,
  permissions,
  onLogout,
}: {
  displayName: string;
  permissions: readonly Permission[];
  onLogout: () => void;
}) {
  const navigation = <ConsoleNavigation permissions={permissions} />;
  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[16rem_minmax(0,1fr)]">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-64 border-r bg-card/45 p-5 lg:flex lg:flex-col">
        <ConsoleIdentity displayName={displayName} />
        <div className="mt-7 min-h-0 flex-1 overflow-y-auto">{navigation}</div>
        <div className="mt-5 flex items-center gap-1 border-t pt-4">
          <ThemeToggle />
          <Button
            variant="ghost"
            className="flex-1 justify-start text-muted-foreground"
            onClick={onLogout}
          >
            <LogOut className="size-4" />
            Sign out
          </Button>
        </div>
      </aside>
      <div className="fixed top-3 left-3 z-50 lg:hidden">
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="secondary" size="icon" aria-label="Open navigation">
              <Menu className="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72 p-5">
            <SheetHeader>
              <SheetTitle className="text-left">Realtime Me</SheetTitle>
            </SheetHeader>
            <div className="mt-6">{navigation}</div>
            <Button
              variant="ghost"
              className="mt-8 w-full justify-start text-muted-foreground"
              onClick={onLogout}
            >
              <LogOut className="size-4" />
              Sign out
            </Button>
          </SheetContent>
        </Sheet>
      </div>
      <div className="min-w-0 lg:col-start-2">
        <Outlet />
      </div>
    </div>
  );
}

function ConsoleIdentity({ displayName }: { displayName: string }) {
  return (
    <div className="px-3">
      <p className="text-xs font-semibold tracking-[0.2em] text-primary uppercase">Realtime Me</p>
      <p className="mt-2 truncate text-sm text-muted-foreground">{displayName}</p>
    </div>
  );
}

function ConsoleNavigation({ permissions }: { permissions: readonly Permission[] }) {
  const allowed = new Set(permissions);
  return (
    <nav className="space-y-6" aria-label="Console">
      {navigationGroups.map((group) => {
        const items = group.items.filter(({ permission }) => allowed.has(permission));
        if (items.length === 0) return null;
        return (
          <div key={group.label}>
            <p className="mb-2 px-3 text-[0.65rem] font-semibold tracking-[0.16em] text-muted-foreground uppercase">
              {group.label}
            </p>
            <div className="space-y-1">
              {items.map(({ label, to, icon: Icon }) => (
                <NavLink
                  key={to}
                  to={to}
                  end={to === "/"}
                  className={({ isActive }) =>
                    cn(
                      "flex h-10 items-center gap-3 rounded-lg px-3 text-sm transition-colors",
                      isActive
                        ? "bg-primary/12 text-primary"
                        : "text-muted-foreground hover:bg-accent hover:text-foreground",
                    )
                  }
                >
                  <Icon className="size-4" />
                  {label}
                </NavLink>
              ))}
            </div>
          </div>
        );
      })}
    </nav>
  );
}
