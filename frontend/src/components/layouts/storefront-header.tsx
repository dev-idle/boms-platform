"use client";

import { useCallback, useEffect, useId, useRef, useState } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import {
  CartIcon,
  HeartIcon,
  SearchIcon,
  UserIcon,
} from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";
import { cn } from "@/lib/utils";
import { StorefrontAccountMenu } from "./storefront-account-menu";
import { StorefrontHeaderSearch } from "./storefront-header-search";
import { StorefrontIconButton } from "./storefront-icon-button";
import { useStorefrontSearchPanel } from "./use-storefront-search-panel";

export function StorefrontHeader() {
  const [scrolled, setScrolled] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const searchPanelId = useId();
  const searchPanelRef = useRef<HTMLDivElement>(null);
  const searchToggleRef = useRef<HTMLButtonElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const status = useAuthStore((state) => state.status);

  const closeSearch = useCallback(() => setSearchOpen(false), []);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useStorefrontSearchPanel({
    open: searchOpen,
    onClose: closeSearch,
    panelRef: searchPanelRef,
    toggleRef: searchToggleRef,
    inputRef: searchInputRef,
  });

  return (
    <header
      className={cn(
        "storefront-header sticky top-0 z-50",
        scrolled && "storefront-header-scrolled",
      )}
    >
      <div className="storefront-container">
        <div className="storefront-header-toolbar">
          <BrandLogo className="min-w-0" size="header" />

          <div
            aria-label="Shop actions"
            className="storefront-header-actions"
            role="toolbar"
          >
            <StorefrontIconButton
              ref={searchToggleRef}
              aria-controls={searchOpen ? searchPanelId : undefined}
              label={searchOpen ? "Close search" : "Search"}
              onClick={() => setSearchOpen((open) => !open)}
              pressed={searchOpen}
            >
              <SearchIcon />
            </StorefrontIconButton>

            <StorefrontIconButton href={ROUTE.products} label="Favorites">
              <HeartIcon />
            </StorefrontIconButton>

            <StorefrontIconButton href={ROUTE.cart} label="Cart">
              <CartIcon />
            </StorefrontIconButton>

            {status === "authenticated" ? (
              <StorefrontAccountMenu />
            ) : (
              <StorefrontIconButton href={ROUTE.login} label="Sign in">
                <UserIcon />
              </StorefrontIconButton>
            )}
          </div>
        </div>
      </div>

      <div
        className={cn(
          "storefront-header-search-wrap",
          searchOpen && "storefront-header-search-wrap--open",
        )}
      >
        <div className="storefront-header-search-wrap__inner">
          <StorefrontHeaderSearch
            inputRef={searchInputRef}
            onClose={closeSearch}
            open={searchOpen}
            panelId={searchPanelId}
            panelRef={searchPanelRef}
          />
        </div>
      </div>
    </header>
  );
}
