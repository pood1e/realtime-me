import { ExternalLink, Images, LayoutGrid, Library, Music } from 'lucide-react';
import type { ReactElement } from 'react';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { EmptyCard, StatusSection } from '@/components/layout';

type ExternalApp = {
  title: string;
  description: string;
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
    title: 'Music',
    description: 'Search and listen. Playlists are kept in your own browser.',
    icon: <Music className="size-5" />,
    path: '/music',
  },
  {
    title: 'Library',
    description: 'The shelf of published books, read in the browser.',
    icon: <Library className="size-5" />,
    path: '/books',
  },
  {
    title: 'Wallpapers',
    description: 'A public gallery. Browse it, take what you like.',
    icon: <Images className="size-5" />,
    path: '/wallpapers',
  },
];

export function AppsApp() {
  const origin = appUrl(import.meta.env.VITE_COMMONS_APP_URL);
  const apps: LinkedApp[] = origin
    ? APPS.map((app) => ({ ...app, href: `${origin}${app.path}` }))
    : [];

  return (
    <StatusSection title="Apps" icon={<LayoutGrid className="size-4" />}>
      {apps.length === 0 ? <EmptyCard text="No apps are configured." /> : apps.map((app) => <AppCard key={app.title} app={app} />)}
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
            <ExternalLink className="size-4 text-muted-foreground transition-colors group-hover:text-primary" aria-hidden />
          </CardAction>
          <CardTitle className="flex items-center gap-2.5 text-lg">
            <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">{app.icon}</span>
            {app.title}
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3">
          <CardDescription>{app.description}</CardDescription>
          <span className="text-xs text-muted-foreground">Opens in a new tab</span>
        </CardContent>
      </Card>
    </a>
  );
}

function appUrl(configured: string | undefined): string | null {
  return configured?.trim().replace(/\/+$/, '') || null;
}
