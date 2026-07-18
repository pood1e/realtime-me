import { AppShell } from "@realtime-me/status-web/components/AppShell";
import { createStatusApi } from "@realtime-me/status-web/lib/transport";
import { lazy, type ReactNode, Suspense } from "react";
import { Route, Routes } from "react-router-dom";
import { CONSOLE_URL, STATUS_API_BASE } from "@/config";

const statusApi = createStatusApi(STATUS_API_BASE);
const ProjectsApp = lazy(() =>
  import("@realtime-me/status-web/routes/Projects").then(({ ProjectsApp }) => ({
    default: ProjectsApp,
  })),
);
const PublicStatusApp = lazy(() =>
  import("@realtime-me/status-web/routes/PublicStatus").then(({ PublicStatusApp }) => ({
    default: PublicStatusApp,
  })),
);
const SharePage = lazy(() =>
  import("@/features/share/SharePage").then(({ SharePage }) => ({ default: SharePage })),
);
const WallpapersPage = lazy(() =>
  import("@/features/wallpapers/WallpapersPage").then(({ WallpapersPage }) => ({
    default: WallpapersPage,
  })),
);

export function App() {
  return (
    <Routes>
      <Route element={<AppShell api={statusApi} consoleUrl={CONSOLE_URL} />}>
        <Route path="/" element={<DeferredPage page={<PublicStatusApp />} />} />
        <Route path="/projects" element={<DeferredPage page={<ProjectsApp />} />} />
        <Route path="/wallpapers" element={<DeferredPage page={<WallpapersPage />} />} />
        <Route path="*" element={<DeferredPage page={<PublicStatusApp />} />} />
      </Route>
      <Route path="/s/:token" element={<DeferredPage page={<SharePage />} />} />
      <Route path="/share/:token" element={<DeferredPage page={<SharePage />} />} />
    </Routes>
  );
}

function DeferredPage({ page }: { page: ReactNode }) {
  return (
    <Suspense
      fallback={<p className="py-16 text-center text-sm text-muted-foreground">Loading…</p>}
    >
      {page}
    </Suspense>
  );
}
