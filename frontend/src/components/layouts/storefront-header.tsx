"use client";

import { useRouter } from "next/navigation";
import {
  useCallback,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
  type RefObject,
} from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import { CloseIcon, MenuIcon, SearchIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { buildCatalogBrowseHref } from "@/features/catalog";

import type { CatalogCategory } from "@/lib/schemas/catalog";
import { StorefrontHeaderAccount } from "./storefront-header-account";
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
  const [panel, setPanel] = useState<HeaderPanel>(null);
  const searchPanelId = useId();
  const menuPanelId = useId();
  const searchPanelRef = useRef<HTMLDivElement>(null);
  const searchToggleRef = useRef<HTMLButtonElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const menuPanelRef = useRef<HTMLDivElement>(null);
  const menuFirstLinkRef = useRef<HTMLAnchorElement>(null);
  // The phone menu button below 64rem, the Categories link from there.
  const menuToggleRef = useRef<HTMLButtonElement>(null);
  const menuTriggerRef = useRef<HTMLButtonElement>(null);
  const menuOpenerRef = useRef<HTMLButtonElement | null>(null);
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

  function toggleMenu(opener: RefObject<HTMLButtonElement | null>): void {
    menuOpenerRef.current = opener.current;
    setPanel((current) => (current === "categories" ? null : "categories"));
  }

  // Closed from inside, the focus goes back to the button that opened the menu.
  function closeMenu(): void {
    if (menuPanelRef.current?.contains(document.activeElement)) {
      menuOpenerRef.current?.focus();
    }
    closePanel();
  }

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
    onClose: closeMenu,
    panelRef: menuPanelRef,
    toggleRef: menuOpenerRef,
    focusRef: menuFirstLinkRef,
  });

  return (
    <header className="storefront-header sticky top-0 z-50">
      <div className="storefront-container">
        <div className="storefront-header-toolbar">
          <StorefrontIconButton
            ref={menuToggleRef}
            aria-controls={menuOpen ? menuPanelId : undefined}
            className="storefront-header-menu-toggle"
            expanded={menuOpen}
            label={menuOpen ? "Close menu" : "Open menu"}
            onClick={() => toggleMenu(menuToggleRef)}
          >
            {menuOpen ? <CloseIcon /> : <MenuIcon />}
          </StorefrontIconButton>
          <BrandLogo size="header" />

          {/* Nav and search share one slot: opening search swaps the row in place. */}
          <div className="storefront-header-center">
            <div
              className="storefront-header-center__nav"
              inert={searchOpen || undefined}
            >
              <StorefrontHeaderNav
                menuOpen={menuOpen}
                menuPanelId={menuPanelId}
                onToggleMenu={() => toggleMenu(menuTriggerRef)}
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
            role="group"
          >
            <StorefrontIconButton
              ref={searchToggleRef}
              aria-controls={searchOpen ? searchPanelId : undefined}
              expanded={searchOpen}
              label={searchOpen ? "Close search" : "Search"}
              onClick={toggleSearch}
            >
              {searchOpen ? <CloseIcon /> : <SearchIcon />}
            </StorefrontIconButton>
            <StorefrontHeaderCartLink />
            <span aria-hidden="true" className="storefront-header-actions__rule" />
            <StorefrontHeaderAccount />
          </div>
        </div>
      </div>

      <StorefrontHeaderCategories
        categories={categories}
        firstLinkRef={menuFirstLinkRef}
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
