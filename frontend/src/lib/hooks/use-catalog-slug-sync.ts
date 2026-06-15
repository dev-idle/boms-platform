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

  function writeField(name: FieldPath<T>, value: string): void {
    setValue(name, value as T[FieldPath<T>], {
      shouldDirty: true,
      shouldValidate: false,
    });
  }

  function handleNameChange(value: string): void {
    writeField("name" as FieldPath<T>, value);
    if (!slugManualRef.current) {
      writeField("slug" as FieldPath<T>, slugifyCatalogName(value));
    }
  }

  function handleSlugChange(value: string): void {
    slugManualRef.current = true;
    writeField("slug" as FieldPath<T>, value);
  }

  return { handleNameChange, handleSlugChange };
}
