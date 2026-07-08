import { Route, Routes } from 'react-router-dom';
import { AppShell } from '@/components/AppShell';
import { InternalStatusApp } from '@/routes/Internal';
import { ProfileApp } from '@/routes/Profile';
import { PublicStatusApp } from '@/routes/PublicStatus';

export function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/" element={<PublicStatusApp />} />
        <Route path="/about" element={<ProfileApp />} />
        <Route path="*" element={<PublicStatusApp />} />
      </Route>
      <Route path="/internal" element={<InternalStatusApp />} />
    </Routes>
  );
}
