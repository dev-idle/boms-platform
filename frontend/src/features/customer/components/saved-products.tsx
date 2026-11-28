"use client";

import { StorefrontAccountSection } from "@/components/layouts/storefront-account-layout";
import { StorefrontEmptyState, StorefrontErrorState } from "@/components/layouts/storefront-state";

import { useSavedProducts } from "../hooks";
import type { SavedList } from "../schemas";
import { SavedProductRow } from "./saved-product-row";

const SECTIONS: ReadonlyArray<{ list: SavedList; title: string; lead: string; emptyTitle: string; empty: string }> = [
  {
    list: "favorite",
    title: "Favorites",
    lead: "Products you order again.",
    emptyTitle: "No favorites yet",
    empty: "Select the heart on a product to keep it here.",
  },
  {
    list: "wishlist",
    title: "Wishlist",
    lead: "Products to try. Moving one to the cart removes it from this list.",
    emptyTitle: "Nothing on your wishlist",
    empty: "Select the bookmark on a product to save it for later.",
  },
];

/** The customer's favorites and wishlist. */
export function SavedProducts() {
  const savedQuery = useSavedProducts();

  if (savedQuery.isPending) {
    return (
      <div aria-busy="true" className="grid gap-6">
        {[0, 1].map((row) => (
          <div className="flex gap-6" key={row}>
            <span className="catalog-tile catalog-tile--row skeleton" />
            <div className="grid flex-1 content-start gap-3">
              <span className="skeleton block h-3 w-24" />
              <span className="skeleton block h-6 w-1/2" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (savedQuery.isError) {
    return <StorefrontErrorState onRetry={() => void savedQuery.refetch()} title="We could not load your saved lists" />;
  }

  return (
    <div className="grid gap-6">
      {SECTIONS.map(({ list, title, lead, emptyTitle, empty }) => {
        const items = savedQuery.data.filter((item) => item.list === list);
        return (
          <StorefrontAccountSection description={lead} id={`saved-${list}`} key={list} title={title}>
            {items.length === 0 ? (
              <StorefrontEmptyState message={empty} title={emptyTitle} />
            ) : (
              <ul className="storefront-saved-lines">
                {items.map((item) => (
                  <SavedProductRow item={item} key={item.product.id} />
                ))}
              </ul>
            )}
          </StorefrontAccountSection>
        );
      })}
    </div>
  );
}
