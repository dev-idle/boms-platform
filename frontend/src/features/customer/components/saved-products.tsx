"use client";

import { Button } from "@/components/ui/button";
import { InlineLoadingState } from "@/components/ui/loading-state";

import { useSavedProducts } from "../hooks";
import type { SavedList } from "../schemas";
import { SavedProductRow } from "./saved-product-row";

const SECTIONS: ReadonlyArray<{ list: SavedList; title: string; lead: string; empty: string }> = [
  {
    list: "favorite",
    title: "Favorites",
    lead: "The ones you come back for, ready to order again.",
    empty: "Select the heart on a product to keep it here.",
  },
  {
    list: "wishlist",
    title: "Wishlist",
    lead: "What you want to try; moving one to the cart takes it off this list.",
    empty: "Select the bookmark on a product to save it for later.",
  },
];

/** The customer's favorites and wishlist. */
export function SavedProducts() {
  const savedQuery = useSavedProducts();

  if (savedQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (savedQuery.isError) {
    return (
      <div className="storefront-empty-state">
        <p className="storefront-empty-state__message text-error">We could not load your lists.</p>
        <Button type="button" variant="outline" onClick={() => void savedQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <div>
      {SECTIONS.map(({ list, title, lead, empty }) => {
        const items = savedQuery.data.filter((item) => item.list === list);
        return (
          <section aria-labelledby={`saved-${list}`} className="storefront-account-section" key={list}>
            <header className="storefront-account-section__header">
              <h2 className="storefront-account-section__title" id={`saved-${list}`}>
                {title}
              </h2>
              <p className="storefront-account-section__lead">{lead}</p>
            </header>
            {items.length === 0 ? (
              <div className="storefront-empty-state storefront-empty-state--card">
                <p className="storefront-empty-state__message">{empty}</p>
              </div>
            ) : (
              <ul className="storefront-cart__lines">
                {items.map((item) => (
                  <SavedProductRow item={item} key={item.product.id} />
                ))}
              </ul>
            )}
          </section>
        );
      })}
    </div>
  );
}
