import { AuthGuard } from "@realtime-me/library-web";
import { API_BASE, AUTH_ORIGIN } from "./config";
import { ImagesPage } from "./ImagesPage";
export function App() {
  return (
    <AuthGuard apiBase={API_BASE} authOrigin={AUTH_ORIGIN}>
      <ImagesPage />
    </AuthGuard>
  );
}
