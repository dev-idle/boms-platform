"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

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
          for (const item of mapValidationDetailsToFormErrors(error.details!)) {
            const field = item.field as keyof CategoryFormInput;
            if (field in values) {
              form.setError(field, { message: item.message });
            }
          }
          return;
        }
        if (error.code === "slug_exists") {
          toast.error("That slug is already in use.");
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
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Name">
                <Input placeholder="Breads" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="slug"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Slug">
                <Input placeholder="breads" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="sort_order"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Sort order">
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
            <FormItem className="flex items-center gap-2">
              <FormControl>
                <input
                  checked={field.value}
                  className="h-4 w-4 rounded border-border"
                  onChange={(event) => field.onChange(event.target.checked)}
                  type="checkbox"
                />
              </FormControl>
              <FormLabel className="!mt-0">Active</FormLabel>
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
