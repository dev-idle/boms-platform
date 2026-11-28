"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";
import { CATALOG_COMBOS_PAGE_SIZE } from "@/constants/catalog";
import { plural } from "@/lib/utils";

import { useCatalogCombos } from "../hooks";
import {
  CATALOG_COMBOS_HEADING_ID,
  CATALOG_COMBOS_SECTION_ID,
  resolveStorefrontCombosHeadingHash,
  scrollToStorefrontCombosSection,
} from "../lib/scroll-to-storefront-anchor";
import { CatalogPagination } from "./catalog-pagination";
import { ComboCard } from "./combo-card";

type ComboCatalogProps = {
  renderPurchaseActions?: (comboId: string) => ReactNode;
};

/** The combo bundles under the products: their own heading, the same cards. */
export function ComboCatalog({ renderPurchaseActions }: ComboCatalogProps) {
  const [page, setPage] = useState(1);
  const filter = useMemo(
    () => ({ page, page_size: CATALOG_COMBOS_PAGE_SIZE }),
    [page],
  );
  const combosQuery = useCatalogCombos(filter);
  const combos = combosQuery.data?.combos ?? [];
  const pagination = combosQuery.data?.pagination;
  const sectionMounted =
    combosQuery.isPending || combosQuery.isError || combos.length > 0;

  useEffect(() => {
    if (!sectionMounted) {
      return;
    }

    const anchorId = resolveStorefrontCombosHeadingHash();
    if (!anchorId) {
      return;
    }

    requestAnimationFrame(() => scrollToStorefrontCombosSection());
  }, [sectionMounted, combosQuery.isPending]);

  // The section shows only once there are bundles to offer.
  if (combosQuery.isPending || (combosQuery.isSuccess && combos.length === 0)) {
    return null;
  }

  return (
    <section
      aria-labelledby={CATALOG_COMBOS_HEADING_ID}
      className="catalog-combos"
      id={CATALOG_COMBOS_SECTION_ID}
    >
      <header className="catalog-combos__header">
        <div className="catalog-combos__copy">
          <h2
            className="catalog-section__title catalog-combos__title"
            id={CATALOG_COMBOS_HEADING_ID}
          >
            Combos
          </h2>
          <p className="catalog-combos__lead">
            Curated sets at a combo price.
          </p>
        </div>
        {pagination ? (
          <p className="catalog-page__count">{plural(pagination.total, "combo")}</p>
        ) : null}
      </header>

      {combosQuery.isError ? (
        <StorefrontErrorState
          onRetry={() => void combosQuery.refetch()}
          title="We could not load the combos"
        />
      ) : (
        <>
          <div className="catalog-product-grid">
            {combos.map((combo) => (
              <ComboCard
                key={combo.id}
                combo={combo}
                renderPurchaseActions={renderPurchaseActions}
              />
            ))}
          </div>

          {pagination ? (
            <CatalogPagination
              onPageChange={setPage}
              page={page}
              totalPages={pagination.total_pages}
            />
          ) : null}
        </>
      )}
    </section>
  );
}
