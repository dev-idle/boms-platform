"use client";

import { useMemo, useState } from "react";

import { ProductPurchaseActions } from "@/features/customer";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogCombos } from "../hooks";
import { CatalogPagination } from "./catalog-pagination";

const COMBOS_PAGE_SIZE = 12;

export function ComboCatalog() {
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
        <p className="text-sm text-muted">Loading combo deals…</p>
      </section>
    );
  }

  if (combosQuery.isError) {
    return (
      <section className="mx-auto w-full max-w-7xl space-y-2 border-t border-border px-4 pb-10 pt-10 sm:px-6 lg:px-8">
        <h2 className="font-heading text-2xl font-medium text-foreground">
          Combo deals
        </h2>
        <p className="text-sm text-error">Failed to load combo deals.</p>
      </section>
    );
  }

  if (combos.length === 0) {
    return null;
  }

  return (
    <section className="mx-auto w-full max-w-7xl space-y-6 border-t border-border px-4 pb-10 pt-10 sm:px-6 lg:px-8">
      <div>
        <h2 className="font-heading text-2xl font-medium text-foreground">
          Combo deals
        </h2>
        <p className="mt-2 text-sm text-muted">
          Limited-time bundles with special pricing.
        </p>
      </div>

      <div className="grid gap-6 sm:grid-cols-2">
        {combos.map((combo) => (
          <article
            key={combo.id}
            className="rounded-lg border border-border bg-surface p-6"
          >
            <h3 className="font-heading text-xl font-medium text-foreground">
              {combo.name}
            </h3>
            <p className="mt-2 text-sm text-muted">
              Valid until {formatDateTime(combo.ends_at)}
            </p>
            <ul className="mt-4 space-y-1 text-sm text-muted">
              {combo.items.map((item) => (
                <li key={`${combo.id}-${item.product_id}`}>
                  {item.quantity}× {item.product_name}
                </li>
              ))}
            </ul>
            <p className="mt-4 text-lg font-medium text-foreground">
              {formatPriceCents(combo.price_cents)}
            </p>
            <div className="mt-4">
              <ProductPurchaseActions
                comboId={combo.id}
                label="Add combo to cart"
              />
            </div>
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
