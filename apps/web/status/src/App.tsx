import { Route, Routes } from 'react-router-dom';
import { AppShell } from '@/components/AppShell';
import { AppsApp } from '@/routes/Apps';
import { InternalStatusApp } from '@/routes/Internal';
import { ProjectsApp } from '@/routes/Projects';
import { PublicStatusApp } from '@/routes/PublicStatus';

export function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/" element={<PublicStatusApp />} />
        <Route path="/apps" element={<AppsApp />} />
        <Route path="/projects" element={<ProjectsApp />} />
        <Route path="*" element={<PublicStatusApp />} />
      </Route>
      <Route path="/internal" element={<InternalStatusApp />} />
    </Routes>
  );
}
