import { Route, Routes } from 'react-router-dom';
import { InternalStatusApp } from '@/routes/Internal';
import { ProfileApp } from '@/routes/Profile';
import { PublicStatusApp } from '@/routes/PublicStatus';

export function App() {
  return (
    <Routes>
      <Route path="/" element={<PublicStatusApp />} />
      <Route path="/about" element={<ProfileApp />} />
      <Route path="/internal" element={<InternalStatusApp />} />
      <Route path="*" element={<PublicStatusApp />} />
    </Routes>
  );
}
