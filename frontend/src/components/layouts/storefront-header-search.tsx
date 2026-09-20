"use client";

import type { KeyboardEvent, RefObject } from "react";

import { CloseIcon, SearchIcon } from "@/components/icons/storefront-icons";
import { CATALOG_SEARCH_MAX_LENGTH } from "@/features/catalog";
import { cn } from "@/lib/utils";

type StorefrontHeaderSearchProps = {
  /** Id of the highlighted suggestion, announced through aria-activedescendant. */
  activeOptionId?: string;
  inputRef: RefObject<HTMLInputElement | null>;
  listboxId: string;
  onClose: () => void;
  onKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
  onQueryChange: (query: string) => void;
  onSubmit: () => void;
  open: boolean;
  panelId: string;
  panelRef: RefObject<HTMLDivElement | null>;
  query: string;
  suggestionsOpen: boolean;
};

/** Search field that takes the nav's place in the header row — no second row, no layout shift. */
export function StorefrontHeaderSearch({
  activeOptionId,
  inputRef,
  listboxId,
  onClose,
  onKeyDown,
  onQueryChange,
  onSubmit,
  open,
  panelId,
  panelRef,
  query,
  suggestionsOpen,
}: StorefrontHeaderSearchProps) {
  return (
    <div
      ref={panelRef}
      className={cn(
        "storefront-header-search",
        open && "storefront-header-search--open",
      )}
      id={panelId}
      inert={!open || undefined}
    >
      <form
        className="search-field"
        onSubmit={(event) => {
          event.preventDefault();
          onSubmit();
        }}
        role="search"
      >
        <div className="search-field__box">
          <span aria-hidden className="search-field__leading">
            <SearchIcon className="search-field__glyph" />
          </span>
          <input
            ref={inputRef}
            aria-activedescendant={activeOptionId}
            aria-autocomplete="list"
            aria-controls={suggestionsOpen ? listboxId : undefined}
            aria-expanded={suggestionsOpen}
            aria-label="Search products"
            autoComplete="off"
            className="search-field__input"
            maxLength={CATALOG_SEARCH_MAX_LENGTH}
            name="search"
            onChange={(event) => onQueryChange(event.target.value)}
            onKeyDown={onKeyDown}
            placeholder="Search pastries, cakes, gifts…"
            role="combobox"
            spellCheck={false}
            type="search"
            value={query}
          />
          {/* Always reachable: below 48rem the field covers the toolbar toggle. */}
          <button
            aria-label="Close search"
            className="search-field__action"
            onClick={onClose}
            type="button"
          >
            <CloseIcon className="size-3.5" />
          </button>
        </div>
      </form>
    </div>
  );
}
