import { Suspense, type ReactNode } from "react";

import { StorefrontCategoryHeader } from "./storefront-category-header";
import { StorefrontFooter } from "./storefront-footer";
import { StorefrontHeader } from "./storefront-header";
import { StorefrontTopbar } from "./storefront-topbar";

type StorefrontShellProps = {
  children: ReactNode;
};

/** Shared chrome for public + customer storefront routes. */
export function StorefrontShell({ children }: StorefrontShellProps) {
  return (
    <div className="flex min-h-full flex-col">
      <StorefrontTopbar />
      <Suspense fallback={<StorefrontHeader categories={[]} />}>
        <StorefrontCategoryHeader />
      </Suspense>
      <main className="flex flex-1 flex-col">{children}</main>
      <StorefrontFooter />
    </div>
  );
}
