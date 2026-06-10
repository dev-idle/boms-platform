import type { ReactNode } from "react";

import { AuthScrollReset } from "./auth-scroll-reset";

type StorefrontAuthShellProps = {
  children: ReactNode;
};

/** Full-bleed auth layout — logo and exit live inside the page shell. */
export function StorefrontAuthShell({ children }: StorefrontAuthShellProps) {
  return (
    <div className="auth-shell flex min-h-dvh flex-col">
      <AuthScrollReset />
      <main className="flex min-h-dvh flex-1 flex-col">{children}</main>
    </div>
  );
}
