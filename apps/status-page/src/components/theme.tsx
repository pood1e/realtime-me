import { useCallback, useEffect, useState } from 'react';
import { Moon, Sun } from 'lucide-react';
import { Button } from '@/components/ui/button';

type Theme = 'light' | 'dark';

const ONE_YEAR_SECONDS = 60 * 60 * 24 * 365;

// The preference lives in a cookie rather than in localStorage, and that is the whole point:
// localStorage belongs to one origin, and the sites on this domain are several. A cookie
// written against the registrable domain is read by all of them, so choosing dark here is
// choosing it everywhere — which is what somebody who picked a theme meant.
//
// Nothing here names the domain. It is derived from the address the page was served from, so
// no real hostname has to live in the source and nothing breaks the first time one moves.
function readTheme(): Theme {
  const stored = document.cookie.match(/(?:^|;\s*)theme=(dark|light)/);
  if (stored?.[1] === 'dark' || stored?.[1] === 'light') return stored[1];
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function writeTheme(theme: Theme): void {
  const attributes = [`theme=${theme}`, 'path=/', `max-age=${ONE_YEAR_SECONDS}`, 'samesite=lax'];
  const host = window.location.hostname;
  const labels = host.split('.');
  if (!/^[\d.]+$/.test(host) && labels.length >= 3) {
    attributes.push(`domain=.${labels.slice(-2).join('.')}`);
  }
  if (window.location.protocol === 'https:') attributes.push('secure');
  document.cookie = attributes.join('; ');
}

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>(readTheme);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
    document.documentElement.classList.toggle('light', theme === 'light');
    writeTheme(theme);
  }, [theme]);

  // Coming back to a tab that was open while the theme was changed on a sibling site. Nothing
  // tells a page that a cookie moved, so it is read again when the page is looked at again —
  // which is the moment it would otherwise be visibly wrong.
  useEffect(() => {
    const resync = () => {
      if (document.visibilityState === 'visible') setTheme(readTheme());
    };
    document.addEventListener('visibilitychange', resync);
    return () => document.removeEventListener('visibilitychange', resync);
  }, []);

  const isDark = theme === 'dark';
  const toggle = useCallback(() => setTheme(isDark ? 'light' : 'dark'), [isDark]);
  return (
    <Button
      variant="ghost"
      size="icon"
      aria-label={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
      title="Toggle theme"
      onClick={toggle}
    >
      {isDark ? <Sun /> : <Moon />}
    </Button>
  );
}
