import { AuthGuard } from "@realtime-me/library-web";
import { ImagesPage } from "./ImagesPage";
import { API_BASE, AUTH_ORIGIN } from "./config";
export function App() {
  return (
    <AuthGuard apiBase={API_BASE} authOrigin={AUTH_ORIGIN}>
      <ImagesPage />
    </AuthGuard>
  );
}
