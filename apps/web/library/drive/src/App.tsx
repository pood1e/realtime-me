import { AuthGuard } from "@realtime-me/library-web";
import { API_BASE, AUTH_ORIGIN } from "./config";
import { DrivePage } from "./DrivePage";

export function App() {
  return (
    <AuthGuard apiBase={API_BASE} authOrigin={AUTH_ORIGIN}>
      <DrivePage />
    </AuthGuard>
  );
}
