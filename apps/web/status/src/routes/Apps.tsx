import { Card, CardAction, CardHeader, CardTitle } from "@realtime-me/web-ui/card";
import { ExternalLink, Images, LayoutDashboard, LayoutGrid, Library, Music } from "lucide-react";
import { type ReactElement, useEffect, useState } from "react";
import { EmptyCard, StatusSection } from "@/components/layout";

type ExternalApp = {
  title: string;
  icon: ReactElement;
  path: string;
};

type LinkedApp = ExternalApp & { href: string };

// The three pages of one public site, which lives in its own project. This page is a launcher
// and nothing else: it holds no contract with that site and fetches nothing from it. Its
// address is build-time configuration, and a build that was not given one shows no links at
// all rather than links that go nowhere.
const APPS: ExternalApp[] = [
  {
    title: "Music",
    icon: <Music className="size-5" />,
    path: "/music",
  },
  {
    title: "Library",
    icon: <Library className="size-5" />,
    path: "/books",
  },
  {
    title: "Wallpapers",
    icon: <Images className="size-5" />,
    path: "/wallpapers",
  },
];

// The console is the owner's own, and only the owner has any use for it, so it is offered only
// to a browser that has recently signed into it. The console says so itself, by leaving a cookie
// on this domain — this page asks it nothing and cannot: the session is a __Host- cookie bound
// to the API's own host, and the alternative to a note left in the open would have been giving
// a page that anyone can open the right to make credentialed calls to the private API.
//
// The note is a hint about what to draw, and nothing more. Forging it earns a visitor a link,
// and following the link shows them the sign-in screen, which is what they would have got
// anyway.
function signedIntoConsole(): boolean {
  return /(?:^|;\s*)signed_in=1/.test(document.cookie);
}

export function AppsApp() {
  const commons = appUrl(import.meta.env.VITE_COMMONS_APP_URL);
  const console_ = appUrl(import.meta.env.VITE_CONSOLE_APP_URL);
  const [signedIn, setSignedIn] = useState(signedIntoConsole);

  // Signing out happens on the other site, and nothing tells this page that a cookie moved. It
  // is read again when the page is looked at again — the moment it would otherwise be offering
  // a door that is no longer there.
  useEffect(() => {
    const resync = () => {
      if (document.visibilityState === "visible") setSignedIn(signedIntoConsole());
    };
    document.addEventListener("visibilitychange", resync);
    return () => document.removeEventListener("visibilitychange", resync);
  }, []);

  const apps: LinkedApp[] = [
    ...(commons ? APPS.map((app) => ({ ...app, href: `${commons}${app.path}` })) : []),
    ...(signedIn && console_
      ? [
          {
            title: "Console",
            icon: <LayoutDashboard className="size-5" />,
            path: "/drive",
            href: `${console_}/drive`,
          },
        ]
      : []),
  ];

  return (
    <StatusSection title="Apps" icon={<LayoutGrid className="size-4" />}>
      {apps.length === 0 ? (
        <EmptyCard text="No apps are configured." />
      ) : (
        apps.map((app) => <AppCard key={app.title} app={app} />)
      )}
    </StatusSection>
  );
}

function AppCard({ app }: { app: LinkedApp }) {
  return (
    <a
      href={app.href}
      target="_blank"
      rel="noreferrer"
      className="group rounded-xl outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
    >
      <Card className="h-full transition-shadow group-hover:ring-primary/40">
        <CardHeader>
          <CardAction>
            <ExternalLink
              className="size-4 text-muted-foreground transition-colors group-hover:text-primary"
              aria-hidden
            />
          </CardAction>
          <CardTitle className="flex items-center gap-2.5 text-lg">
            <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
              {app.icon}
            </span>
            {app.title}
          </CardTitle>
        </CardHeader>
      </Card>
    </a>
  );
}

function appUrl(configured: string | undefined): string | null {
  return configured?.trim().replace(/\/+$/, "") || null;
}
