"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";

import { CatalogImageListField } from "@/components/ui/catalog-image-list-field";
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
import { MoneyFieldInput } from "@/components/ui/money-field-input";
import { isCloudinaryConfigured } from "@/lib/cloudinary/config";
import { cloudinaryComboImageFieldHint } from "@/lib/cloudinary/messages";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";
import { windowEndsInDays, windowStartsNow } from "@/lib/validation/datetime";

import {
  useCreateCombo,
  useUpdateCombo,
} from "../hooks";
import {
  comboFormSchema,
  type ComboFormInput,
  type ComboFormValues,
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

export function ComboForm({ mode, combo, onSuccess }: ComboFormProps) {
  const createCombo = useCreateCombo();
  const updateCombo = useUpdateCombo(combo?.id ?? "");

  const defaultValues = useMemo(
    (): ComboFormValues => ({
      name: combo?.name ?? "",
      slug: combo?.slug ?? "",
      price_cents: combo?.price_cents,
      image_url: combo?.image_url ?? "",
      starts_at: combo?.starts_at ?? windowStartsNow(),
      ends_at: combo?.ends_at ?? windowEndsInDays(7),
      is_active: combo?.is_active ?? true,
      items:
        combo?.items.map((item) => ({
          product_id: item.product_id,
          quantity: item.quantity,
        })) ?? [],
    }),
    [combo],
  );

  const form = useForm<ComboFormValues, unknown, ComboFormInput>({
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
              <FieldControl hint={FORM_FIELD_HINT.catalogPrice} label="Price">
                <MoneyFieldInput {...field} />
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
                    ? cloudinaryComboImageFieldHint()
                    : FORM_FIELD_HINT.comboImageUrlFallback
                }
                label="Combo image"
                optional
              >
                <CatalogImageListField
                  disabled={isSavePending}
                  maxImages={1}
                  onChange={(next) => field.onChange(next[0] ?? "")}
                  value={field.value ? [field.value] : []}
                />
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
