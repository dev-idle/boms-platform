"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { CatalogNameSlugFields } from "@/components/ui/catalog-name-slug-fields";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import {
  Form,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { FormPublishSwitch } from "@/components/ui/form-publish-switch";
import { Input } from "@/components/ui/input";
import {
  FORM_FIELD_HINT,
  FORM_SWITCH_HINT,
  FORM_SWITCH_LABEL,
} from "@/constants/dashboard-form-copy";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useCreateCategory, useUpdateCategory } from "../hooks";
import {
  categoryFormSchema,
  type CategoryFormInput,
  type ManagerCategory,
} from "../schemas";

type CategoryFormProps = {
  mode: "create" | "edit";
  category?: ManagerCategory;
  onSuccess?: () => void;
};

const CATEGORY_FORM_FIELDS = [
  "name",
  "slug",
  "sort_order",
  "is_active",
] as const;

export function CategoryForm({ mode, category, onSuccess }: CategoryFormProps) {
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory(category?.id ?? "");

  const defaultValues: CategoryFormInput = {
    name: category?.name ?? "",
    slug: category?.slug ?? "",
    sort_order: category?.sort_order ?? 0,
    is_active: category?.is_active ?? true,
  };

  const form = useForm<CategoryFormInput>({
    resolver: zodResolver(categoryFormSchema),
    defaultValues,
  });

  function onSubmit(values: CategoryFormInput): void {
    const mutation = mode === "create" ? createCategory : updateCategory;
    mutation.mutate(values, {
      onSuccess: () => onSuccess?.(),
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to save category");
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, CATEGORY_FORM_FIELDS);
          return;
        }
        if (error.code === "slug_exists") {
          form.setError("slug", { message: "This slug is already in use" });
          return;
        }
        toast.error(error.message);
      },
    });
  }

  const isPending = createCategory.isPending || updateCategory.isPending;

  return (
    <Form {...form}>
      <form
        className="dashboard-profile-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <CatalogNameSlugFields
          control={form.control}
          mode={mode}
          namePlaceholder="Breads"
          setValue={form.setValue}
          slugPlaceholder="breads"
        />

        <FormField
          control={form.control}
          name="sort_order"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_FIELD_HINT.catalogSortOrder}
                hintId="category-sort-order-hint"
                label="Sort order"
              >
                <Input min={0} type="number" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="is_active"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_SWITCH_HINT.storefrontVisible}
                hintId="category-storefront-visible-hint"
                label={FORM_SWITCH_LABEL.storefrontVisible}
                variant="switch"
              >
                <FormPublishSwitch
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel={mode === "create" ? "Create category" : "Save changes"}
            isPending={isPending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
