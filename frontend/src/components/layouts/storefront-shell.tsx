import type { ReactNode } from "react";
import { connection } from "next/server";

import {
  dalListCatalogCategories,
  STOREFRONT_CATEGORY_PAGE_SIZE,
} from "@/lib/dal/catalog";
import { StorefrontFooter } from "./storefront-footer";
import { StorefrontHeader } from "./storefront-header";
import { StorefrontTopbar } from "./storefront-topbar";

type StorefrontShellProps = {
  children: ReactNode;
};

/** Shared chrome for public + customer storefront routes. */
export async function StorefrontShell({ children }: StorefrontShellProps) {
  // Header nav is live catalog data — render at request time, never prerender.
  await connection();

  // Nav degrades to "Shop all" + "Combos" if the catalog API is unreachable.
  const categories = await dalListCatalogCategories(
    STOREFRONT_CATEGORY_PAGE_SIZE,
  ).catch(() => []);

  return (
    <div className="flex min-h-full flex-col">
      <StorefrontTopbar />
      <StorefrontHeader categories={categories} />
      <main className="flex flex-1 flex-col">{children}</main>
      <StorefrontFooter />
    </div>
  );
}
