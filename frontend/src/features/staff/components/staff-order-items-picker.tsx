"use client";

import { useState } from "react";

import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { IntegerFieldInput } from "@/components/ui/integer-field-input";
import { Label } from "@/components/ui/label";
import { useCatalogCombos, useCatalogProducts } from "@/features/catalog";
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";
import { formatPriceCents } from "@/lib/validation/catalog";

/** A line of an order being taken, priced as the catalog shows it. */
export type StaffOrderLine = {
  /** The product or combo id. */
  id: string;
  kind: "product" | "combo";
  name: string;
  unitPriceCents: number;
  quantity: number;
  soldOutToday: boolean;
};

type StaffOrderItemsPickerProps = {
  lines: StaffOrderLine[];
  onChange: (lines: StaffOrderLine[]) => void;
  disabled: boolean;
};

/** Matches the dashboard table search: long enough to skip a keystroke burst. */
const SEARCH_DEBOUNCE_MS = 280;
const PRODUCT_RESULTS = 8;
/** Every combo on sale: the catalog lists at most this many a page, and has no combo search. */
const COMBOS_ON_SALE = 100;

/**
 * Finds products by name and the combos on sale, and holds the lines chosen.
 * A custom cake is configured online, so it is not offered here.
 */
export function StaffOrderItemsPicker({ lines, onChange, disabled }: StaffOrderItemsPickerProps) {
  const [search, setSearch] = useState("");
  const [debounced, settling] = useDebouncedValue(search.trim(), SEARCH_DEBOUNCE_MS);
  const productsQuery = useCatalogProducts({ page: 1, page_size: PRODUCT_RESULTS, search: debounced });
  const combosQuery = useCatalogCombos({ page: 1, page_size: COMBOS_ON_SALE });
  const chosen = new Set(lines.map((line) => line.id));
  // Every search, not just the first: the products of the previous term stay
  // on screen while the next ones load.
  const busy = settling || productsQuery.isFetching || combosQuery.isFetching;
  const failed = productsQuery.isError || combosQuery.isError;
  const offers: StaffOrderLine[] = [
    ...(productsQuery.data?.products ?? [])
      .filter((product) => !product.is_customizable)
      .map((product) => ({
        id: product.id,
        kind: "product" as const,
        name: product.name,
        unitPriceCents: product.price_cents,
        quantity: 1,
        soldOutToday: product.sold_out_today,
      })),
    ...(combosQuery.data?.combos ?? [])
      .filter((combo) => combo.name.toLowerCase().includes(debounced.toLowerCase()))
      .map((combo) => ({
        id: combo.id,
        kind: "combo" as const,
        name: combo.name,
        unitPriceCents: combo.price_cents,
        quantity: 1,
        soldOutToday: combo.sold_out_today,
      })),
  ];

  function setQuantity(id: string, quantity: number | undefined): void {
    onChange(lines.map((line) => (line.id === id ? { ...line, quantity: quantity ?? 1 } : line)));
  }

  return (
    <div className="staff-order-items">
      <div className="staff-order-items__field">
        <Label htmlFor="staff-order-search">Find a product or combo</Label>
        <DashboardSearchField
          id="staff-order-search"
          onChange={setSearch}
          onClear={() => setSearch("")}
          placeholder="Search by name"
          value={search}
        />
      </div>

      <ul aria-busy={busy || undefined} aria-label="Products and combos" className="staff-order-items__list">
        {failed ? (
          <li className="staff-order-items__row staff-order-items__row--state text-error">
            <span>The catalog could not be loaded.</span>
            <DashboardTableActionButton
              label="Load the catalog again"
              onClick={() => {
                void productsQuery.refetch();
                void combosQuery.refetch();
              }}
              text="Try again"
              tone="accent"
            />
          </li>
        ) : busy ? (
          <li className="staff-order-items__row staff-order-items__row--state text-muted">Searching…</li>
        ) : offers.length === 0 ? (
          <li className="staff-order-items__row staff-order-items__row--state text-muted">Nothing matches that name.</li>
        ) : (
          offers.map((offer) => (
            <li key={offer.id} className="staff-order-items__row">
              <span className="staff-order-items__name">
                {offer.name}
                <span className="staff-order-items__meta">
                  {offer.kind === "combo" ? "Combo" : null}
                  {offer.soldOutToday ? <span className="staff-order-items__sold-out">Sold out today</span> : null}
                </span>
              </span>
              <span className="staff-order-items__price">{formatPriceCents(offer.unitPriceCents)}</span>
              <DashboardTableActionButton
                blockedReason={disabled ? "Taking the order" : chosen.has(offer.id) ? "Already on the order" : undefined}
                label={`Add ${offer.name}`}
                onClick={() => onChange([...lines, offer])}
                text="Add"
                tone="accent"
              />
            </li>
          ))
        )}
      </ul>

      {lines.length > 0 ? (
        <section aria-labelledby="staff-order-chosen" className="staff-order-items__field">
          <p className="text-form-label" id="staff-order-chosen">
            On the order
          </p>
          <ul className="staff-order-items__list staff-order-items__list--chosen">
            {lines.map((line) => (
              <li key={line.id} className="staff-order-items__row">
                <span className="staff-order-items__name">{line.name}</span>
                <IntegerFieldInput
                  aria-label={`How many ${line.name}`}
                  disabled={disabled}
                  max={99}
                  min={1}
                  value={line.quantity}
                  onChange={(quantity) => setQuantity(line.id, quantity)}
                />
                <span className="staff-order-items__price">{formatPriceCents(line.unitPriceCents * line.quantity)}</span>
                <DashboardTableActionButton
                  blockedReason={disabled ? "Taking the order" : undefined}
                  label={`Remove ${line.name}`}
                  onClick={() => onChange(lines.filter((other) => other.id !== line.id))}
                  text="Remove"
                  tone="danger"
                />
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
