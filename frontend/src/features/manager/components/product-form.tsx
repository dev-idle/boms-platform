"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { usePathname } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { CatalogNameSlugFields } from "@/components/ui/catalog-name-slug-fields";
import { CatalogImageField } from "@/components/ui/catalog-image-field";
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
import { Select } from "@/components/ui/select";
import { isCloudinaryConfigured } from "@/lib/cloudinary/config";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useCategories, useCreateProduct, useUpdateProduct } from "../hooks";
import {
  productFormSchema,
  type ManagerProduct,
  type ProductFormInput,
} from "../schemas";

const CREATE_PRODUCT_EMPTY_VALUES: ProductFormInput = {
  category_id: "",
  name: "",
  slug: "",
  description: null,
  price_cents: 0,
  is_available: true,
  image_url: null,
};

type ProductFormProps = {
  mode: "create" | "edit";
  product?: ManagerProduct;
  onSuccess?: () => void;
};

const PRODUCT_FORM_FIELDS = [
  "category_id",
  "name",
  "slug",
  "description",
  "price_cents",
  "is_available",
  "image_url",
] as const;

export function ProductForm({ mode, product, onSuccess }: ProductFormProps) {
  const pathname = usePathname();
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
  const resetForm = form.reset;

  useEffect(() => {
    if (mode === "create") {
      resetForm(CREATE_PRODUCT_EMPTY_VALUES);
    }
  }, [mode, pathname, resetForm]);

  function onSubmit(values: ProductFormInput): void {
    const mutation = mode === "create" ? createProduct : updateProduct;
    mutation.mutate(values, {
      onSuccess: () => {
        if (mode === "create") {
          form.reset(CREATE_PRODUCT_EMPTY_VALUES);
        }
        onSuccess?.();
      },
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to save product");
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, PRODUCT_FORM_FIELDS);
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
                <Select onChange={field.onChange} value={field.value ?? ""}>
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
        <CatalogNameSlugFields
          key={mode === "create" ? pathname : product?.id}
          control={form.control}
          mode={mode}
          namePlaceholder="Sourdough loaf"
          setValue={form.setValue}
          slugPlaceholder="sourdough-loaf"
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
              <FieldControl
                hint={FORM_FIELD_HINT.catalogPriceCents}
                label="Price (cents)"
              >
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
              <FieldControl
                hint={
                  isCloudinaryConfigured()
                    ? FORM_FIELD_HINT.productImageCloudinary
                    : FORM_FIELD_HINT.productImageUrlFallback
                }
                label="Product image"
                optional
              >
                {isCloudinaryConfigured() ? (
                  <CatalogImageField
                    disabled={isPending}
                    onChange={field.onChange}
                    value={field.value ?? null}
                  />
                ) : (
                  <Input
                    placeholder="https://example.com/image.jpg"
                    value={field.value ?? ""}
                    onChange={(event) =>
                      field.onChange(event.target.value || null)
                    }
                  />
                )}
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="is_available"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_SWITCH_HINT.availableToOrder}
                hintId="product-available-to-order-hint"
                label={FORM_SWITCH_LABEL.availableToOrder}
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
            idleLabel={mode === "create" ? "Create product" : "Save changes"}
            isPending={isPending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
