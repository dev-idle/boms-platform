"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { FieldControl } from "@/components/ui/field-control";
import { Label } from "@/components/ui/label";
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
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";
import { applyFormFieldErrors } from "@/lib/validation";
import {
  fromDatetimeLocalValue,
  toDatetimeLocalValue,
} from "@/lib/validation/datetime";

import {
  useCreateCombo,
  useProducts,
  useUpdateCombo,
} from "../hooks";
import {
  comboFormSchema,
  type ComboFormInput,
  type ManagerCombo,
} from "../schemas";

type ComboFormProps = {
  mode: "create" | "edit";
  combo?: ManagerCombo;
  onSuccess?: () => void;
};

const COMBO_FORM_FIELDS = [
  "name",
  "slug",
  "price_cents",
  "starts_at",
  "ends_at",
  "is_active",
  "items",
] as const;

const defaultStartsAt = (): string =>
  new Date(Date.now() + 60 * 60 * 1000).toISOString();
const defaultEndsAt = (): string =>
  new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();

export function ComboForm({ mode, combo, onSuccess }: ComboFormProps) {
  const createCombo = useCreateCombo();
  const updateCombo = useUpdateCombo(combo?.id ?? "");
  const [productSearchInput, setProductSearchInput] = useState("");
  const productSearch = useDebouncedValue(productSearchInput, 280).trim();
  const productsQuery = useProducts({
    page: 1,
    page_size: 100,
    search: productSearch,
    category_id: "",
  });
  const productOptions = useMemo(() => {
    const options = new Map<string, { id: string; name: string }>();
    for (const product of productsQuery.data?.products ?? []) {
      options.set(product.id, { id: product.id, name: product.name });
    }
    if (combo) {
      for (const item of combo.items) {
        if (!options.has(item.product_id)) {
          options.set(item.product_id, {
            id: item.product_id,
            name: item.product_name,
          });
        }
      }
    }
    return Array.from(options.values()).sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [combo, productsQuery.data?.products]);

  const defaultValues = useMemo(
    (): ComboFormInput => ({
      name: combo?.name ?? "",
      slug: combo?.slug ?? "",
      price_cents: combo?.price_cents ?? 0,
      starts_at: combo?.starts_at ?? defaultStartsAt(),
      ends_at: combo?.ends_at ?? defaultEndsAt(),
      is_active: combo?.is_active ?? true,
      items: combo?.items.map((item) => ({
        product_id: item.product_id,
        quantity: item.quantity,
      })) ?? [{ product_id: "", quantity: 1 }],
    }),
    [combo],
  );

  const form = useForm<ComboFormInput>({
    resolver: zodResolver(comboFormSchema),
    defaultValues,
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "items",
  });

  function onSubmit(values: ComboFormInput): void {
    const mutation = mode === "create" ? createCombo : updateCombo;
    mutation.mutate(values, {
      onSuccess: () => onSuccess?.(),
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Something went wrong");
          return;
        }
        if (error.isSlugExists()) {
          form.setError("slug", { message: "This slug is already in use" });
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, COMBO_FORM_FIELDS);
          return;
        }
        toast.error(error.message);
      },
    });
  }

  const isSavePending = createCombo.isPending || updateCombo.isPending;

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
                <Input {...field} />
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
                <Input {...field} />
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
              <FieldControl label="Combo price (cents)">
                <Input min={0} step={1} type="number" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="starts_at"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Starts at">
                  <Input
                    type="datetime-local"
                    value={toDatetimeLocalValue(field.value)}
                    onChange={(event) =>
                      field.onChange(fromDatetimeLocalValue(event.target.value))
                    }
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="ends_at"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Ends at">
                  <Input
                    type="datetime-local"
                    value={toDatetimeLocalValue(field.value)}
                    onChange={(event) =>
                      field.onChange(fromDatetimeLocalValue(event.target.value))
                    }
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <FormField
          control={form.control}
          name="is_active"
          render={({ field }) => (
            <FormItem className="flex items-center gap-2">
              <FormControl>
                <input
                  checked={field.value}
                  className="size-4 rounded border-border"
                  onChange={(event) => field.onChange(event.target.checked)}
                  type="checkbox"
                />
              </FormControl>
              <FormLabel className="!mt-0">Active</FormLabel>
            </FormItem>
          )}
        />

        <div className="space-y-3">
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div className="min-w-[16rem] flex-1 space-y-2">
              <Label className="text-form-label">Products in combo</Label>
              <DashboardSearchField
                onChange={setProductSearchInput}
                onClear={() => setProductSearchInput("")}
                placeholder="Search products by name or slug"
                value={productSearchInput}
              />
            </div>
            <Button
              onClick={() => append({ product_id: "", quantity: 1 })}
              type="button"
              variant="outline"
            >
              Add product
            </Button>
          </div>
          {productsQuery.isPending ? (
            <p className="text-sm text-muted">Loading products…</p>
          ) : null}
          {fields.map((field, index) => (
            <div
              key={field.id}
              className="grid gap-3 rounded-card border border-border p-3 sm:grid-cols-[1fr_120px_auto]"
            >
              <FormField
                control={form.control}
                name={`items.${index}.product_id`}
                render={({ field: productField }) => (
                  <FormItem>
                    <FormLabel className="sr-only">Product</FormLabel>
                    <FormControl>
                      <Select
                        onChange={productField.onChange}
                        value={productField.value}
                      >
                        <option value="">Select product</option>
                        {productOptions.map((product) => (
                          <option key={product.id} value={product.id}>
                            {product.name}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name={`items.${index}.quantity`}
                render={({ field: quantityField }) => (
                  <FormItem>
                    <FormLabel className="sr-only">Quantity</FormLabel>
                    <FormControl>
                      <Input
                        min={1}
                        step={1}
                        type="number"
                        {...quantityField}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                disabled={fields.length <= 1}
                onClick={() => remove(index)}
                type="button"
                variant="outline"
              >
                Remove
              </Button>
            </div>
          ))}
        </div>

        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel={mode === "create" ? "Create combo" : "Save changes"}
            isPending={isSavePending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
