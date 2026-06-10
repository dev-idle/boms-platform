"use client";

import Link from "next/link";
import { Suspense, useEffect, useState } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import { CatalogSearchForm } from "@/features/catalog/components/catalog-search-form";
import { ROUTE } from "@/constants/routes";
import { cn } from "@/lib/utils";

import { StorefrontAccountLink } from "./storefront-account-link";

export function StorefrontHeader() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={cn(
        "sticky top-0 z-50 bg-bg/95 backdrop-blur-sm transition-[border-color] duration-standard ease-default",
        scrolled && "border-b border-border",
      )}
    >
      <div className="mx-auto flex h-14 max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
        <BrandLogo />

        <nav
          aria-label="Storefront"
          className="hidden flex-1 items-center justify-center gap-8 text-sm font-medium text-muted md:flex"
        >
          <Link
            className="flex min-h-11 items-center rounded-full px-3 py-2 transition-colors duration-standard ease-default hover:text-rose-500"
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
            className="flex min-h-11 items-center rounded-full px-3 py-2 text-sm font-medium text-muted transition-colors duration-standard ease-default hover:bg-blush hover:text-ink md:hidden"
            href={ROUTE.products}
          >
            Search
          </Link>
          <Link
            className="flex min-h-11 items-center rounded-full px-3 py-2 text-sm font-medium text-muted transition-colors duration-standard ease-default hover:bg-blush hover:text-ink"
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
