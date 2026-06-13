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
import { Select } from "@/components/ui/select";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useCategories, useCreateProduct, useUpdateProduct } from "../hooks";
import {
  productFormSchema,
  type ManagerProduct,
  type ProductFormInput,
} from "../schemas";

type ProductFormProps = {
  mode: "create" | "edit";
  product?: ManagerProduct;
  onSuccess?: () => void;
};

export function ProductForm({ mode, product, onSuccess }: ProductFormProps) {
  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct(product?.id ?? "");
  const categoriesQuery = useCategories({
    page: 1,
    page_size: 100,
    search: "",
  });
  const categories = categoriesQuery.data?.categories ?? [];

  const defaultValues: ProductFormInput = {
    category_id: product?.category_id ?? "",
    name: product?.name ?? "",
    slug: product?.slug ?? "",
    description: product?.description ?? null,
    price_cents: product?.price_cents ?? 0,
    is_available: product?.is_available ?? true,
    image_url: product?.image_url ?? null,
  };

  const form = useForm<ProductFormInput>({
    resolver: zodResolver(productFormSchema),
    defaultValues,
  });

  function onSubmit(values: ProductFormInput): void {
    const mutation = mode === "create" ? createProduct : updateProduct;
    mutation.mutate(values, {
      onSuccess: () => onSuccess?.(),
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to save product");
          return;
        }
        if (error.hasValidationDetails()) {
          for (const item of mapValidationDetailsToFormErrors(error.details!)) {
            const field = item.field as keyof ProductFormInput;
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

  const isPending = createProduct.isPending || updateProduct.isPending;

  return (
    <Form {...form}>
      <form
        className="dashboard-profile-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="category_id"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Category">
                <Select onChange={field.onChange} value={field.value}>
                  <option value="">Select a category</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </Select>
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Name">
                <Input placeholder="Sourdough loaf" {...field} />
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
                <Input placeholder="sourdough-loaf" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Description" optional>
                <Input
                  placeholder="Short product description"
                  value={field.value ?? ""}
                  onChange={(event) =>
                    field.onChange(event.target.value || null)
                  }
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="price_cents"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Price (cents)">
                <Input min={0} type="number" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="image_url"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Image URL" optional>
                <Input
                  placeholder="https://example.com/image.jpg"
                  value={field.value ?? ""}
                  onChange={(event) =>
                    field.onChange(event.target.value || null)
                  }
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="is_available"
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
              <FormLabel className="!mt-0">Available</FormLabel>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel={mode === "create" ? "Create product" : "Save changes"}
            isPending={isPending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
