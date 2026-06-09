import Link from "next/link";
import { Suspense } from "react";

import { CatalogSearchForm } from "@/features/catalog";
import { ROUTE } from "@/constants/routes";

import { StorefrontAccountLink } from "./storefront-account-link";

export function StorefrontHeader() {
  return (
    <header className="sticky top-0 z-50 border-b border-border bg-surface/95 backdrop-blur-sm">
      <div className="mx-auto flex h-14 max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
        <Link
          className="shrink-0 font-heading text-lg font-medium tracking-tight text-foreground transition-colors duration-default ease-default hover:text-primary"
          href={ROUTE.home}
        >
          BOMS
        </Link>

        <nav
          aria-label="Storefront"
          className="hidden flex-1 items-center justify-center gap-8 text-sm font-medium text-muted md:flex"
        >
          <Link
            className="transition-colors duration-default ease-default hover:text-foreground"
            href={ROUTE.products}
          >
            Shop
          </Link>
        </nav>

        <div className="ml-auto flex min-w-0 flex-1 items-center justify-end gap-1 sm:gap-2 md:flex-none">
          <Suspense fallback={null}>
            <CatalogSearchForm
              className="hidden min-w-0 flex-1 md:block md:max-w-xs lg:max-w-sm"
              inputClassName="h-9"
              showSubmitButton={false}
            />
          </Suspense>
          <Link
            className="rounded-md px-3 py-2 text-sm font-medium text-muted transition-colors duration-default ease-default hover:bg-surface-alt hover:text-foreground md:hidden"
            href={ROUTE.products}
          >
            Search
          </Link>
          <Link
            className="rounded-md px-3 py-2 text-sm font-medium text-muted transition-colors duration-default ease-default hover:bg-surface-alt hover:text-foreground"
            href={ROUTE.cart}
          >
            Cart
          </Link>
          <StorefrontAccountLink />
        </div>
      </div>
    </header>
  );
}
