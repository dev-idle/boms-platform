"use client";

import { useEffect, useId, useMemo, useRef, useState } from "react";

import { DashboardCloseIcon } from "@/components/icons/dashboard-ui-icons";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";
import { cn } from "@/lib/utils";

import { useProducts } from "../hooks";

export type ComboProductPickerValue = {
  id: string;
  name: string;
} | null;

type ComboProductPickerProps = {
  className?: string;
  excludedProductIds: ReadonlySet<string>;
  onChange: (value: ComboProductPickerValue) => void;
  value: ComboProductPickerValue;
};

/** Search-and-pick product control for combo bundle composer. */
export function ComboProductPicker({
  className,
  excludedProductIds,
  onChange,
  value,
}: ComboProductPickerProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listboxId = useId();
  const inputId = useId();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const debouncedQuery = useDebouncedValue(query, 280).trim();

  const productsQuery = useProducts({
    page: 1,
    page_size: 100,
    search: debouncedQuery,
    category_id: "",
  });

  const options = useMemo(() => {
    const items =
      productsQuery.data?.products
        .filter((product) => !excludedProductIds.has(product.id))
        .map((product) => ({ id: product.id, name: product.name })) ?? [];
    return items.sort((a, b) => a.name.localeCompare(b.name));
  }, [excludedProductIds, productsQuery.data?.products]);

  useEffect(() => {
    function handlePointerDown(event: MouseEvent): void {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, []);

  function clearSelection(): void {
    onChange(null);
    setQuery("");
    setOpen(false);
    inputRef.current?.focus();
  }

  function selectProduct(id: string, name: string): void {
    onChange({ id, name });
    setQuery("");
    setOpen(false);
  }

  const showList = open && !value;
  const listEmpty = !productsQuery.isPending && options.length === 0;

  return (
    <div className={cn("combo-product-picker", className)} ref={rootRef}>
      {value ? (
        <div className="combo-product-picker__selected">
          <span className="combo-product-picker__selected-name">{value.name}</span>
          <button
            aria-label={`Clear ${value.name}`}
            className="combo-product-picker__clear"
            onClick={clearSelection}
            type="button"
          >
            <DashboardCloseIcon className="size-3.5" />
          </button>
        </div>
      ) : (
        <div className="combo-product-picker__field">
          <Input
            aria-autocomplete="list"
            aria-controls={listboxId}
            aria-expanded={showList}
            autoComplete="off"
            className="combo-product-picker__input"
            id={inputId}
            onChange={(event) => {
              setQuery(event.target.value);
              setOpen(true);
            }}
            onFocus={() => setOpen(true)}
            placeholder="Add a product…"
            ref={inputRef}
            role="combobox"
            spellCheck={false}
            type="search"
            value={query}
            variant="inline"
          />
          <button
            aria-label="Clear search"
            className={cn(
              "combo-product-picker__search-clear",
              !query && "is-empty",
            )}
            onClick={() => {
              setQuery("");
              setOpen(true);
              inputRef.current?.focus();
            }}
            tabIndex={query ? 0 : -1}
            type="button"
          >
            <DashboardCloseIcon className="size-3.5" />
          </button>
          {showList ? (
            <ul
              className="combo-product-picker__list"
              id={listboxId}
              role="listbox"
            >
              {productsQuery.isPending ? (
                <li className="combo-product-picker__status">Searching…</li>
              ) : null}
              {!productsQuery.isPending && listEmpty ? (
                <li className="combo-product-picker__status">
                  {debouncedQuery
                    ? "No products match."
                    : "Type a product name or slug."}
                </li>
              ) : null}
              {!productsQuery.isPending
                ? options.map((product) => (
                    <li key={product.id} role="presentation">
                      <button
                        aria-selected={false}
                        className="combo-product-picker__option"
                        onClick={() => selectProduct(product.id, product.name)}
                        role="option"
                        type="button"
                      >
                        {product.name}
                      </button>
                    </li>
                  ))
                : null}
            </ul>
          ) : null}
        </div>
      )}
    </div>
  );
}
