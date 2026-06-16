"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";

import { CatalogNameSlugFields } from "@/components/ui/catalog-name-slug-fields";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { FormPublishSwitch } from "@/components/ui/form-publish-switch";
import {
  FORM_FIELD_HINT,
  FORM_SWITCH_HINT,
  FORM_SWITCH_LABEL,
} from "@/constants/dashboard-form-copy";
import {
  Form,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { DashboardDatetimeInput } from "@/components/ui/dashboard-datetime-input";
import { IntegerFieldInput } from "@/components/ui/integer-field-input";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import {
  useCreateCombo,
  useUpdateCombo,
} from "../hooks";
import {
  comboFormSchema,
  type ComboFormInput,
  type ManagerCombo,
} from "../schemas";

import { ComboFormItemsSection } from "./combo-form-items-section";

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

  const defaultValues = useMemo(
    (): ComboFormInput => ({
      name: combo?.name ?? "",
      slug: combo?.slug ?? "",
      price_cents: combo?.price_cents ?? 0,
      starts_at: combo?.starts_at ?? defaultStartsAt(),
      ends_at: combo?.ends_at ?? defaultEndsAt(),
      is_active: combo?.is_active ?? true,
      items:
        combo?.items.map((item) => ({
          product_id: item.product_id,
          quantity: item.quantity,
        })) ?? [],
    }),
    [combo],
  );

  const form = useForm<ComboFormInput>({
    resolver: zodResolver(comboFormSchema),
    defaultValues,
  });

  const { append, fields, remove } = useFieldArray({
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
        <CatalogNameSlugFields
          control={form.control}
          mode={mode}
          namePlaceholder="Weekend pastry box"
          setValue={form.setValue}
          slugPlaceholder="weekend-pastry-box"
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
                <IntegerFieldInput min={0} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="dashboard-datetime-grid grid gap-4 sm:grid-cols-2">
          <FormField
            control={form.control}
            name="starts_at"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Starts at">
                  <DashboardDatetimeInput
                    onChange={field.onChange}
                    value={field.value}
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
                  <DashboardDatetimeInput
                    onChange={field.onChange}
                    value={field.value}
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <ComboFormItemsSection
          append={append}
          combo={combo}
          control={form.control}
          fields={fields}
          remove={remove}
        />

        <FormField
          control={form.control}
          name="is_active"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_SWITCH_HINT.storefrontVisible}
                hintId="combo-storefront-visible-hint"
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
            idleLabel={mode === "create" ? "Create combo" : "Save changes"}
            isPending={isSavePending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
