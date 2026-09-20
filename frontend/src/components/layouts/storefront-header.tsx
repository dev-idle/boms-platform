"use client";

import { useRouter } from "next/navigation";
import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import { SearchIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { buildCatalogBrowseHref } from "@/features/catalog";
import { cn } from "@/lib/utils";

import type { CatalogCategory } from "@/lib/schemas/catalog";
import { StorefrontHeaderAccountLink } from "./storefront-header-account-link";
import { StorefrontHeaderCartLink } from "./storefront-header-cart-link";
import { StorefrontHeaderCategories } from "./storefront-header-categories";
import { StorefrontHeaderNav } from "./storefront-header-nav";
import { StorefrontHeaderSearch } from "./storefront-header-search";
import { StorefrontHeaderSearchResults } from "./storefront-header-search-results";
import { StorefrontIconButton } from "./storefront-icon-button";
import { useStorefrontHeaderPanel } from "./use-storefront-header-panel";
import { useStorefrontSearchSuggestions } from "./use-storefront-search-suggestions";

type StorefrontHeaderProps = {
  categories: CatalogCategory[];
};

/** Only one header panel is open at a time, so the toolbar never shows two states at once. */
type HeaderPanel = "search" | "categories" | null;

export function StorefrontHeader({ categories }: StorefrontHeaderProps) {
  const [scrolled, setScrolled] = useState(false);
  const [panel, setPanel] = useState<HeaderPanel>(null);
  const searchPanelId = useId();
  const menuPanelId = useId();
  const searchPanelRef = useRef<HTMLDivElement>(null);
  const searchToggleRef = useRef<HTMLButtonElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const menuPanelRef = useRef<HTMLDivElement>(null);
  const menuTriggerRef = useRef<HTMLButtonElement>(null);
  const resultsPanelRef = useRef<HTMLDivElement>(null);
  const listboxId = useId();
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(-1);

  const searchOpen = panel === "search";
  const menuOpen = panel === "categories";
  const suggestions = useStorefrontSearchSuggestions(query, searchOpen);
  const suggestionsOpen = searchOpen && suggestions.active;
  // Every product row plus the trailing "see everything" row.
  const optionCount = suggestions.products.length + 1;
  // A shrinking result list must not leave the highlight pointing past the end.
  const activeOption = activeIndex < optionCount ? activeIndex : -1;
  const optionId = (index: number) => `${listboxId}-option-${index}`;

  const closePanel = useCallback(() => {
    setPanel(null);
    setActiveIndex(-1);
  }, []);

  function toggleSearch(): void {
    setPanel((current) => (current === "search" ? null : "search"));
    setActiveIndex(-1);
  }

  function submitSearch(): void {
    const search = query.trim();
    if (!search) {
      return;
    }
    router.push(buildCatalogBrowseHref({ search, page: 1 }));
    closePanel();
  }

  function handleSearchKeyDown(event: KeyboardEvent<HTMLInputElement>): void {
    if (!suggestionsOpen) {
      return;
    }

    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      const next = activeOption + (event.key === "ArrowDown" ? 1 : -1);
      setActiveIndex(
        next < 0 ? optionCount - 1 : next >= optionCount ? 0 : next,
      );
      return;
    }

    if (event.key === "Enter" && activeOption >= 0) {
      const product = suggestions.products[activeOption];
      if (product) {
        event.preventDefault();
        router.push(ROUTE.productDetail(product.id));
        closePanel();
      }
    }
  }

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useStorefrontHeaderPanel({
    open: searchOpen,
    onClose: closePanel,
    panelRef: searchPanelRef,
    toggleRef: searchToggleRef,
    focusRef: searchInputRef,
    satelliteRef: resultsPanelRef,
  });

  useStorefrontHeaderPanel({
    open: menuOpen,
    onClose: closePanel,
    panelRef: menuPanelRef,
    toggleRef: menuTriggerRef,
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

          {/* Nav and search share one slot: opening search swaps the row in place. */}
          <div className="storefront-header-center">
            <div
              className="storefront-header-center__nav"
              inert={searchOpen || undefined}
            >
              <StorefrontHeaderNav
                menuOpen={menuOpen}
                menuPanelId={menuPanelId}
                onToggleMenu={() =>
                  setPanel((current) =>
                    current === "categories" ? null : "categories",
                  )
                }
                triggerRef={menuTriggerRef}
              />
            </div>
            <StorefrontHeaderSearch
              activeOptionId={
                activeOption >= 0 ? optionId(activeOption) : undefined
              }
              inputRef={searchInputRef}
              listboxId={listboxId}
              onClose={closePanel}
              onKeyDown={handleSearchKeyDown}
              onQueryChange={(next) => {
                setQuery(next);
                setActiveIndex(-1);
              }}
              onSubmit={submitSearch}
              open={searchOpen}
              panelId={searchPanelId}
              panelRef={searchPanelRef}
              query={query}
              suggestionsOpen={suggestionsOpen}
            />
          </div>

          <div
            aria-label="Shop actions"
            className="storefront-header-actions"
            role="toolbar"
          >
            <StorefrontIconButton
              ref={searchToggleRef}
              aria-controls={searchOpen ? searchPanelId : undefined}
              label={searchOpen ? "Close search" : "Search"}
              onClick={toggleSearch}
              pressed={searchOpen}
            >
              <SearchIcon />
            </StorefrontIconButton>

            <StorefrontHeaderAccountLink />
            <StorefrontHeaderCartLink />
          </div>
        </div>
      </div>

      <StorefrontHeaderCategories
        categories={categories}
        onNavigate={closePanel}
        open={menuOpen}
        panelId={menuPanelId}
        panelRef={menuPanelRef}
      />

      <StorefrontHeaderSearchResults
        activeIndex={activeOption}
        listboxId={listboxId}
        loading={suggestions.loading}
        onNavigate={closePanel}
        open={suggestionsOpen}
        optionId={optionId}
        panelRef={resultsPanelRef}
        products={suggestions.products}
        query={suggestions.settledQuery}
        total={suggestions.total}
      />
    </header>
  );
}
