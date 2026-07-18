import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { Permission, type Session, SessionService } from "@realtime-me/auth-contracts";
import { ConsolePage, ConsoleShell } from "@realtime-me/web-shell";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@realtime-me/web-ui";
import { ShieldAlert } from "lucide-react";
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { Navigate } from "react-router-dom";
import { AUTH_API_BASE } from "@/config";

const sessionClient = createClient(
  SessionService,
  createConnectTransport({
    baseUrl: AUTH_API_BASE,
    useBinaryFormat: false,
    defaultTimeoutMs: 15_000,
    fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
  }),
);

const SessionContext = createContext<Session | null>(null);

export function SessionBoundary() {
  const [session, setSession] = useState<Session>();
  const [error, setError] = useState<unknown>();
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    setError(undefined);
    void sessionClient
      .getSession({}, { signal: controller.signal })
      .then((response) => {
        if (!response.session) throw new Error("Console session response is empty");
        setSession(response.session);
      })
      .catch((cause: unknown) => {
        if (controller.signal.aborted) return;
        if (cause instanceof ConnectError && cause.code === Code.Unauthenticated) {
          redirectToLogin();
          return;
        }
        setError(cause);
      });
    return () => controller.abort();
  }, [attempt]);

  const logout = useCallback(async () => {
    try {
      await sessionClient.logout({});
      window.location.assign("/signed-out");
    } catch (cause) {
      setError(cause);
    }
  }, []);

  if (error) {
    return (
      <ConsolePage
        title="Console unavailable"
        subtitle="The authenticated session could not be loaded"
      >
        <Card className="mx-auto mt-16 max-w-xl">
          <CardHeader>
            <CardTitle>Unable to open the console</CardTitle>
            <CardDescription>{message(error)}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => setAttempt((current) => current + 1)}>Retry</Button>
          </CardContent>
        </Card>
      </ConsolePage>
    );
  }
  if (!session) {
    return (
      <div className="grid min-h-dvh place-items-center text-sm text-muted-foreground">
        Loading secure console…
      </div>
    );
  }
  return (
    <SessionContext.Provider value={session}>
      <ConsoleShell
        displayName={session.displayName || session.subject}
        permissions={session.permissions}
        onLogout={() => void logout()}
      />
    </SessionContext.Provider>
  );
}

export function PermissionBoundary({
  permission,
  children,
}: PropsWithChildren<{ permission: Permission }>) {
  const session = useSession();
  if (session.permissions.includes(permission)) return children;
  return (
    <ConsolePage title="Access denied" subtitle="Your identity does not have this capability">
      <Card className="mx-auto mt-16 max-w-xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldAlert className="size-5 text-destructive" />
            Permission required
          </CardTitle>
          <CardDescription>
            The service also enforces this permission at its API boundary.
          </CardDescription>
        </CardHeader>
      </Card>
    </ConsolePage>
  );
}

export function ConsoleHome() {
  const session = useSession();
  const destination = firstDestination(session.permissions);
  if (destination) return <Navigate to={destination} replace />;
  return (
    <ConsolePage title="Console" subtitle="No owner capabilities were assigned">
      <Card className="mx-auto mt-16 max-w-xl">
        <CardHeader>
          <CardTitle>No services available</CardTitle>
          <CardDescription>
            Assign at least one Realtime Me permission in the identity provider.
          </CardDescription>
        </CardHeader>
      </Card>
    </ConsolePage>
  );
}

export function SignedOutPage() {
  return (
    <main className="grid min-h-dvh place-items-center p-6">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Signed out</CardTitle>
          <CardDescription>Your server-side Console session has ended.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild>
            <a href="/auth/login">Sign in again</a>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}

function useSession(): Session {
  const session = useContext(SessionContext);
  if (!session) throw new Error("Session context is unavailable");
  return session;
}

function redirectToLogin() {
  const returnTo = `${window.location.pathname}${window.location.search}`;
  window.location.replace(`/auth/login?return_to=${encodeURIComponent(returnTo)}`);
}

function firstDestination(permissions: readonly Permission[]): string | undefined {
  const destinations: ReadonlyArray<readonly [Permission, string]> = [
    [Permission.STATUS_INTERNAL_READ, "/status"],
    [Permission.LIBRARY_MANAGE, "/library/drive"],
    [Permission.MANAGER_CONTROL, "/manager"],
  ];
  return destinations.find(([permission]) => permissions.includes(permission))?.[1];
}

function message(error: unknown): string {
  return error instanceof Error ? error.message : "Unexpected Console error";
}
