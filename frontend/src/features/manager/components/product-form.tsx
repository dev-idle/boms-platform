"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
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

  const form = useForm<ProductFormInput>({
    resolver: zodResolver(productFormSchema),
    defaultValues: {
      category_id: product?.category_id ?? "",
      name: product?.name ?? "",
      slug: product?.slug ?? "",
      description: product?.description ?? null,
      price_cents: product?.price_cents ?? 0,
      is_available: product?.is_available ?? true,
      image_url: product?.image_url ?? null,
    },
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
      <form className="space-y-4" noValidate onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          control={form.control}
          name="category_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Category</FormLabel>
              <FormControl>
                <select
                  className="flex h-10 w-full rounded-md border border-border bg-surface px-3 py-2 text-sm"
                  onChange={field.onChange}
                  value={field.value}
                >
                  <option value="">Select a category</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="Sourdough loaf" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="slug"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Slug</FormLabel>
              <FormControl>
                <Input placeholder="sourdough-loaf" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="description"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Description</FormLabel>
              <FormControl>
                <Input
                  placeholder="Optional description"
                  value={field.value ?? ""}
                  onChange={(event) =>
                    field.onChange(event.target.value || null)
                  }
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="price_cents"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Price (cents)</FormLabel>
              <FormControl>
                <Input min={0} type="number" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="image_url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Image URL</FormLabel>
              <FormControl>
                <Input
                  placeholder="https://example.com/image.jpg"
                  value={field.value ?? ""}
                  onChange={(event) =>
                    field.onChange(event.target.value || null)
                  }
                />
              </FormControl>
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
        <Button disabled={isPending} type="submit">
          {mode === "create" ? "Create product" : "Save changes"}
        </Button>
      </form>
    </Form>
  );
}
