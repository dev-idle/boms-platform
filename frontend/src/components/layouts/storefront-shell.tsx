import type { ReactNode } from "react";

import { StorefrontFooter } from "./storefront-footer";
import { StorefrontHeader } from "./storefront-header";

type StorefrontShellProps = {
  children: ReactNode;
};

/** Shared chrome for public + customer storefront routes. */
export function StorefrontShell({ children }: StorefrontShellProps) {
  return (
    <div className="flex min-h-full flex-col">
      <StorefrontHeader />
      <main className="flex flex-1 flex-col">{children}</main>
      <StorefrontFooter />
    </div>
  );
}
