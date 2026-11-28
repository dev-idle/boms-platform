import { PageLoadingState } from "@/components/ui/loading-state";

/** The dashboards' whole-page wait for the session, a sign-out or a redirect. */
export function AuthGateShell() {
  return <PageLoadingState />;
}
