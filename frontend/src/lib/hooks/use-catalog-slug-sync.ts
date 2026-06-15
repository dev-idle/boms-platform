"use client";

import { useRef } from "react";
import type { FieldPath, FieldValues, UseFormSetValue } from "react-hook-form";

import { slugifyCatalogName } from "@/lib/validation/catalog";

type CatalogSlugFormValues = {
  name: string;
  slug: string;
};

/**
 * Keeps slug in sync with name on create until the manager edits slug manually.
 * Edit mode never auto-overwrites slug when name changes.
 */
export function useCatalogSlugSync<T extends FieldValues & CatalogSlugFormValues>(
  mode: "create" | "edit",
  setValue: UseFormSetValue<T>,
) {
  const slugManualRef = useRef(mode === "edit");

  function handleNameChange(
    value: string,
    fieldOnChange: (next: string) => void,
  ): void {
    fieldOnChange(value);
    if (!slugManualRef.current) {
      setValue(
        "slug" as FieldPath<T>,
        slugifyCatalogName(value) as T[FieldPath<T>],
        {
          shouldDirty: true,
          shouldValidate: true,
        },
      );
    }
  }

  function handleSlugChange(
    value: string,
    fieldOnChange: (next: string) => void,
  ): void {
    slugManualRef.current = true;
    fieldOnChange(value);
  }

  return { handleNameChange, handleSlugChange };
}
