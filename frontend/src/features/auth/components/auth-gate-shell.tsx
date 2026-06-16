import { LoadingState } from "@/components/ui/loading-state";

/** Auth gate wait shell — single visual for session, sign-out, and redirect. */
export function AuthGateShell() {
  return <LoadingState variant="page" />;
}
