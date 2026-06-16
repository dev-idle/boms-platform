"use client";

import type { ReactNode } from "react";
import { useMemo, useState } from "react";

import { InlineLoadingState } from "@/components/ui/loading-state";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogCombos } from "../hooks";
import { CatalogPagination } from "./catalog-pagination";

const COMBOS_PAGE_SIZE = 12;

type ComboCatalogProps = {
  renderPurchaseActions?: (comboId: string) => ReactNode;
};

export function ComboCatalog({ renderPurchaseActions }: ComboCatalogProps) {
  const [page, setPage] = useState(1);
  const filter = useMemo(
    () => ({ page, page_size: COMBOS_PAGE_SIZE }),
    [page],
  );
  const combosQuery = useCatalogCombos(filter);
  const combos = combosQuery.data?.combos ?? [];
  const pagination = combosQuery.data?.pagination;

  if (combosQuery.isPending) {
    return (
      <section className="mx-auto w-full max-w-7xl px-4 pb-10 pt-10 sm:px-6 lg:px-8">
        <InlineLoadingState />
      </section>
    );
  }

  if (combosQuery.isError) {
    return (
      <section className="mx-auto w-full max-w-7xl space-y-2 px-4 pb-10 pt-10 sm:px-6 lg:px-8">
        <h2 className="text-h2">Combo deals</h2>
        <p className="text-caption text-error">Failed to load combo deals.</p>
      </section>
    );
  }

  if (combos.length === 0) {
    return null;
  }

  return (
    <section className="mx-auto w-full max-w-7xl space-y-6 bg-mint px-4 pb-12 pt-12 sm:px-6 lg:px-8">
      <div>
        <h2 className="text-h2">Combo deals</h2>
        <p className="text-body mt-2">
          Limited-time bundles with special pricing.
        </p>
      </div>

      <div className="grid gap-5 sm:grid-cols-2">
        {combos.map((combo) => (
          <article
            key={combo.id}
            className="rounded-card bg-surface p-6 shadow-rest"
          >
            <h3 className="text-h3">{combo.name}</h3>
            <p className="text-caption mt-2">
              Valid until {formatDateTime(combo.ends_at)}
            </p>
            <ul className="mt-4 space-y-1 text-sm text-muted">
              {combo.items.map((item) => (
                <li key={`${combo.id}-${item.product_id}`}>
                  {item.quantity}× {item.product_name}
                </li>
              ))}
            </ul>
            <p className="text-price mt-4 text-lg">
              {formatPriceCents(combo.price_cents)}
            </p>
            {renderPurchaseActions ? (
              <div className="mt-4">{renderPurchaseActions(combo.id)}</div>
            ) : null}
          </article>
        ))}
      </div>

      {pagination ? (
        <CatalogPagination
          onPageChange={setPage}
          page={page}
          totalPages={pagination.total_pages}
        />
      ) : null}
    </section>
  );
}
