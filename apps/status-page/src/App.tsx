import { InternalStatusApp } from '@/routes/Internal';
import { ProfileApp } from '@/routes/Profile';
import { PublicStatusApp } from '@/routes/PublicStatus';

export function App() {
  const path = window.location.pathname;
  if (path.startsWith('/internal')) return <InternalStatusApp />;
  if (path.startsWith('/about')) return <ProfileApp />;
  return <PublicStatusApp />;
}
