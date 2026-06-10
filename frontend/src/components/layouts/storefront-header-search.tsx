"use client";

import type { RefObject } from "react";
import { useRouter } from "next/navigation";

import { CloseIcon, SearchIcon } from "@/components/icons/storefront-icons";
import {
  buildCatalogBrowseHref,
  CATALOG_SEARCH_MAX_LENGTH,
} from "@/features/catalog/lib/catalog-browse-params";

type StorefrontHeaderSearchProps = {
  inputRef: RefObject<HTMLInputElement | null>;
  onClose: () => void;
  panelRef: RefObject<HTMLDivElement | null>;
  panelId: string;
};

export function StorefrontHeaderSearch({
  inputRef,
  onClose,
  panelRef,
  panelId,
}: StorefrontHeaderSearchProps) {
  const router = useRouter();

  return (
    <div
      ref={panelRef}
      className="storefront-header-search"
      id={panelId}
    >
      <div className="storefront-container">
        <form
          className="storefront-header-search-form"
          onSubmit={(event) => {
            event.preventDefault();
            const formData = new FormData(event.currentTarget);
            const query = String(formData.get("search") ?? "").trim();
            router.push(
              buildCatalogBrowseHref({
                search: query,
                page: 1,
              }),
            );
            onClose();
          }}
          role="search"
        >
          <div className="storefront-header-search-field">
            <SearchIcon className="storefront-header-search-field-icon size-[1.125rem]" />
            <input
              ref={inputRef}
              aria-label="Search products"
              autoComplete="off"
              className="storefront-header-search-input"
              maxLength={CATALOG_SEARCH_MAX_LENGTH}
              name="search"
              placeholder="Search pastries, cakes, gifts…"
              type="search"
            />
            <button
              aria-label="Close search"
              className="storefront-header-search-dismiss"
              onClick={onClose}
              type="button"
            >
              <CloseIcon />
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
