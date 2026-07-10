import { AuthGuard, ToastProvider } from "@cloud-drive/shared";
import { DrivePage } from "./DrivePage";
import { API_BASE, AUTH_ORIGIN } from "./config";

export function App() {
  return (
    <ToastProvider>
      <AuthGuard apiBase={API_BASE} authOrigin={AUTH_ORIGIN}>
        <DrivePage />
      </AuthGuard>
    </ToastProvider>
  );
}
