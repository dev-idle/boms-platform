"use client";

import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useCatalogCombos } from "../hooks";

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
    return null;
  }

  if (combosQuery.isError) {
    return (
      <section className="space-y-2">
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-50">
          Combo deals
        </h2>
        <p className="text-sm text-red-600">Failed to load combo deals.</p>
      </section>
    );
  }

  if (combos.length === 0) {
    return null;
  }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-50">
          Combo deals
        </h2>
        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
          Limited-time bundles with special pricing.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {combos.map((combo) => (
          <article
            key={combo.id}
            className="rounded-lg border border-amber-200 bg-amber-50/50 p-4 dark:border-amber-900/50 dark:bg-amber-950/20"
          >
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
              {combo.name}
            </h3>
            <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
              Valid until {formatDateTime(combo.ends_at)}
            </p>
            <ul className="mt-3 space-y-1 text-sm text-zinc-700 dark:text-zinc-300">
              {combo.items.map((item) => (
                <li key={`${combo.id}-${item.product_id}`}>
                  {item.quantity}× {item.product_name}
                </li>
              ))}
            </ul>
            <p className="mt-4 text-lg font-semibold text-zinc-900 dark:text-zinc-50">
              {formatPriceCents(combo.price_cents)}
            </p>
          </article>
        ))}
      </div>

      {pagination && pagination.total_pages > 1 ? (
        <div className="flex items-center gap-2">
          <Button
            disabled={page <= 1}
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            type="button"
            variant="outline"
          >
            Previous
          </Button>
          <span className="text-sm text-zinc-600">
            Page {pagination.page} of {pagination.total_pages}
          </span>
          <Button
            disabled={page >= pagination.total_pages}
            onClick={() => setPage((current) => current + 1)}
            type="button"
            variant="outline"
          >
            Next
          </Button>
        </div>
      ) : null}
    </section>
  );
}
