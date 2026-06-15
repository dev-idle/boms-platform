"use client";

import { useId } from "react";
import type { Control, FieldPath, FieldValues, UseFormSetValue } from "react-hook-form";

import { FieldControl } from "@/components/ui/field-control";
import {
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { FORM_FIELD_HINT } from "@/constants/dashboard-form-copy";
import { useCatalogSlugSync } from "@/lib/hooks/use-catalog-slug-sync";

type CatalogNameSlugValues = {
  name: string;
  slug: string;
};

type CatalogNameSlugFieldsProps<T extends FieldValues & CatalogNameSlugValues> = {
  control: Control<T>;
  mode: "create" | "edit";
  namePlaceholder?: string;
  setValue: UseFormSetValue<T>;
  slugPlaceholder?: string;
};

export function CatalogNameSlugFields<T extends FieldValues & CatalogNameSlugValues>({
  control,
  mode,
  namePlaceholder = "Display name",
  setValue,
  slugPlaceholder = "url-slug",
}: CatalogNameSlugFieldsProps<T>) {
  const slugHintId = useId();
  const { handleNameChange, handleSlugChange } = useCatalogSlugSync(
    mode,
    setValue,
  );

  return (
    <>
      <FormField
        control={control}
        name={"name" as FieldPath<T>}
        render={({ field }) => (
          <FormItem>
            <FieldControl label="Name">
              <Input
                placeholder={namePlaceholder}
                {...field}
                onChange={(event) =>
                  handleNameChange(event.target.value, field.onChange)
                }
              />
            </FieldControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={control}
        name={"slug" as FieldPath<T>}
        render={({ field }) => (
          <FormItem>
            <FieldControl
              hint={mode === "create" ? FORM_FIELD_HINT.catalogSlugCreate : undefined}
              hintId={slugHintId}
              label="Slug"
            >
              <Input
                placeholder={slugPlaceholder}
                {...field}
                onChange={(event) =>
                  handleSlugChange(event.target.value, field.onChange)
                }
              />
            </FieldControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </>
  );
}
