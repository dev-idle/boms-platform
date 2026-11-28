"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useMemo } from "react";

import { ROUTE } from "@/constants/routes";

import {
  buildCatalogBrowseHref,
  parseCatalogBrowseParams,
  type CatalogBrowseParams,
} from "../lib/catalog-browse-params";

export function useCatalogBrowseFilters() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const params = useMemo(
    () => parseCatalogBrowseParams(searchParams),
    [searchParams],
  );

  const replaceParams = useCallback(
    (next: Partial<CatalogBrowseParams>) => {
      if (pathname !== ROUTE.products) {
        return;
      }

      const merged = {
        ...params,
        ...next,
      };

      router.replace(buildCatalogBrowseHref(merged), { scroll: false });
      // A new list under a new title replaces the one on screen. A page left
      // scrolled into the old list would end up wherever the new, shorter one
      // stops, so it goes back to its top, where the category's title is.
      window.scrollTo({ behavior: "instant", top: 0 });
    },
    [params, pathname, router],
  );

  const setSearch = useCallback(
    (search: string) => {
      replaceParams({ search, page: 1 });
    },
    [replaceParams],
  );

  const setCategory = useCallback(
    (category: string | undefined) => {
      replaceParams({ category, page: 1 });
    },
    [replaceParams],
  );

  const setPage = useCallback(
    (page: number) => {
      replaceParams({ page });
    },
    [replaceParams],
  );

  const clearFilters = useCallback(() => {
    replaceParams({ search: "", category: undefined, page: 1 });
  }, [replaceParams]);

  return {
    params,
    setSearch,
    setCategory,
    setPage,
    clearFilters,
    hasActiveFilters: params.search !== "" || params.category !== undefined,
  };
}
