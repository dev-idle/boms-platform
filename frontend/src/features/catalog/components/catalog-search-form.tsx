"use client";

import type { ReactNode, Ref } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

import {
  buildCatalogBrowseHref,
  CATALOG_SEARCH_MAX_LENGTH,
} from "../lib/catalog-browse-params";

type CatalogSearchFormProps = {
  className?: string;
  inputClassName?: string;
  showSubmitButton?: boolean;
  defaultValue?: string;
  /** When set, updates in-place filters instead of navigating away. */
  onSearch?: (query: string) => void;
  trailing?: ReactNode;
  inputRef?: Ref<HTMLInputElement>;
};

export function CatalogSearchForm({
  className,
  inputClassName,
  showSubmitButton = true,
  defaultValue = "",
  onSearch,
  trailing,
  inputRef,
}: CatalogSearchFormProps) {
  const router = useRouter();

  return (
    <form
      className={className}
      onSubmit={(event) => {
        event.preventDefault();
        const formData = new FormData(event.currentTarget);
        const query = String(formData.get("search") ?? "").trim();
        if (onSearch) {
          onSearch(query);
          return;
        }
        router.push(
          buildCatalogBrowseHref({
            search: query,
            page: 1,
          }),
        );
      }}
      role="search"
    >
      <div className="flex flex-wrap items-center gap-2">
        <Input
          ref={inputRef}
          aria-label="Search products"
          className={inputClassName}
          defaultValue={defaultValue}
          maxLength={CATALOG_SEARCH_MAX_LENGTH}
          name="search"
          placeholder="Search pastries, cakes…"
          type="search"
          variant="inline"
        />
        {showSubmitButton ? <Button type="submit">Search</Button> : null}
        {trailing}
      </div>
    </form>
  );
}
