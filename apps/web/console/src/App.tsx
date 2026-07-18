import { Permission } from "@realtime-me/auth-contracts";
import { createStatusApi } from "@realtime-me/status-web/lib/transport";
import { lazy, type ReactNode, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { ConsoleHome, PermissionBoundary, SessionBoundary, SignedOutPage } from "@/auth/session";
import { STATUS_API_BASE } from "@/config";

const statusApi = createStatusApi(STATUS_API_BASE);
const BooksPage = lazy(() =>
  import("@/features/library/books/BooksPage").then(({ BooksPage }) => ({ default: BooksPage })),
);
const DrivePage = lazy(() =>
  import("@/features/library/drive/DrivePage").then(({ DrivePage }) => ({ default: DrivePage })),
);
const ImagesPage = lazy(() =>
  import("@/features/library/images/ImagesPage").then(({ ImagesPage }) => ({
    default: ImagesPage,
  })),
);
const InternalStatusApp = lazy(() =>
  import("@realtime-me/status-web/routes/Internal").then(({ InternalStatusApp }) => ({
    default: InternalStatusApp,
  })),
);
const ManagerPage = lazy(() =>
  import("@/features/manager/ManagerPage").then(({ ManagerPage }) => ({ default: ManagerPage })),
);
const MusicPage = lazy(() =>
  import("@/features/library/music/MusicPage").then(({ MusicPage }) => ({ default: MusicPage })),
);

export function App() {
  return (
    <Routes>
      <Route path="/signed-out" element={<SignedOutPage />} />
      <Route element={<SessionBoundary />}>
        <Route index element={<ConsoleHome />} />
        <Route
          path="/status"
          element={
            <PermissionBoundary permission={Permission.STATUS_INTERNAL_READ}>
              <DeferredPage page={<InternalStatusApp api={statusApi} />} />
            </PermissionBoundary>
          }
        />
        <Route
          path="/library/drive"
          element={
            <PermissionBoundary permission={Permission.LIBRARY_MANAGE}>
              <DeferredPage page={<DrivePage />} />
            </PermissionBoundary>
          }
        />
        <Route
          path="/library/books"
          element={
            <PermissionBoundary permission={Permission.LIBRARY_MANAGE}>
              <DeferredPage page={<BooksPage />} />
            </PermissionBoundary>
          }
        />
        <Route
          path="/library/music"
          element={
            <PermissionBoundary permission={Permission.LIBRARY_MANAGE}>
              <DeferredPage page={<MusicPage />} />
            </PermissionBoundary>
          }
        />
        <Route
          path="/library/images"
          element={
            <PermissionBoundary permission={Permission.LIBRARY_MANAGE}>
              <DeferredPage page={<ImagesPage />} />
            </PermissionBoundary>
          }
        />
        <Route
          path="/manager"
          element={
            <PermissionBoundary permission={Permission.MANAGER_CONTROL}>
              <DeferredPage page={<ManagerPage />} />
            </PermissionBoundary>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}

function DeferredPage({ page }: { page: ReactNode }) {
  return (
    <Suspense
      fallback={
        <div className="grid min-h-[60dvh] place-items-center text-sm text-muted-foreground">
          Loading…
        </div>
      }
    >
      {page}
    </Suspense>
  );
}
