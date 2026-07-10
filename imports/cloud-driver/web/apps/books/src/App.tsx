import { AuthGuard, ToastProvider } from "@cloud-drive/shared";
import { BooksPage } from "./BooksPage";
import { API_BASE, AUTH_ORIGIN } from "./config";
export function App() {
  return (
    <ToastProvider>
      <AuthGuard apiBase={API_BASE} authOrigin={AUTH_ORIGIN}>
        <BooksPage />
      </AuthGuard>
    </ToastProvider>
  );
}
