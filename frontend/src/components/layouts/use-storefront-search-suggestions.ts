"use client";

import { useCatalogProducts } from "@/features/catalog";
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";
import type { CatalogProduct } from "@/lib/schemas/catalog";

/** Shortest query worth a round trip; one or two letters match almost everything. */
const MIN_QUERY_LENGTH = 2;
const SUGGESTION_LIMIT = 6;
const DEBOUNCE_MS = 220;

type StorefrontSearchSuggestions = {
  products: CatalogProduct[];
  /** Everything the query matches, not just the rows shown. */
  total: number;
  /** True while the catalogue is answering the current query. */
  loading: boolean;
  /** True once the query is long enough to have an answer. */
  active: boolean;
  /** The query the visible results belong to. */
  settledQuery: string;
};

/** Live product matches for the header search field. */
export function useStorefrontSearchSuggestions(
  query: string,
  enabled: boolean,
): StorefrontSearchSuggestions {
  const settledQuery = useDebouncedValue(query, DEBOUNCE_MS).trim();
  const active = enabled && settledQuery.length >= MIN_QUERY_LENGTH;

  const productsQuery = useCatalogProducts(
    { page: 1, page_size: SUGGESTION_LIMIT, search: settledQuery },
    { enabled: active },
  );

  return {
    products: active ? (productsQuery.data?.products ?? []) : [],
    total: active ? (productsQuery.data?.pagination.total ?? 0) : 0,
    loading: active && productsQuery.isFetching,
    active,
    settledQuery,
  };
}
