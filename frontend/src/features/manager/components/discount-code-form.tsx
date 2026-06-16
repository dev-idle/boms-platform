"use client";

import { useMemo } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
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
import { FieldControl } from "@/components/ui/field-control";
import {
  IntegerFieldInput,
  OptionalIntegerFieldInput,
} from "@/components/ui/integer-field-input";
import { DashboardDatetimeInput } from "@/components/ui/dashboard-datetime-input";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import {
  useCreateDiscountCode,
  useUpdateDiscountCode,
} from "../hooks";
import {
  DISCOUNT_TYPE,
  discountCodeFormSchema,
  type DiscountCodeFormInput,
  type DiscountCodeFormValues,
  type ManagerDiscountCode,
} from "../schemas";

type DiscountCodeFormProps = {
  mode: "create" | "edit";
  discountCode?: ManagerDiscountCode;
  onSuccess?: () => void;
};

const defaultStartsAt = (): string =>
  new Date(Date.now() + 60 * 60 * 1000).toISOString();
const defaultEndsAt = (): string =>
  new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString();

const DISCOUNT_CODE_FORM_FIELDS = [
  "code",
  "discount_type",
  "value",
  "min_order_cents",
  "max_uses",
  "starts_at",
  "ends_at",
  "is_active",
] as const;

export function DiscountCodeForm({
  mode,
  discountCode,
  onSuccess,
}: DiscountCodeFormProps) {
  const createDiscountCode = useCreateDiscountCode();
  const updateDiscountCode = useUpdateDiscountCode(discountCode?.id ?? "");

  const defaultValues = useMemo(
    (): DiscountCodeFormValues => ({
      code: discountCode?.code ?? "",
      discount_type: discountCode?.discount_type ?? DISCOUNT_TYPE.percent,
      ...(discountCode ? { value: discountCode.value } : {}),
      min_order_cents: discountCode?.min_order_cents ?? null,
      max_uses: discountCode?.max_uses ?? null,
      starts_at: discountCode?.starts_at ?? defaultStartsAt(),
      ends_at: discountCode?.ends_at ?? defaultEndsAt(),
      is_active: discountCode?.is_active ?? true,
    }),
    [discountCode],
  );

  const form = useForm<DiscountCodeFormValues, unknown, DiscountCodeFormInput>({
    resolver: zodResolver(discountCodeFormSchema),
    defaultValues,
  });

  const discountType = useWatch({
    control: form.control,
    name: "discount_type",
  });

  function onSubmit(values: DiscountCodeFormInput): void {
    const mutation = mode === "create" ? createDiscountCode : updateDiscountCode;
    mutation.mutate(values, {
      onSuccess: () => onSuccess?.(),
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Something went wrong");
          return;
        }
        if (error.isCodeExists()) {
          form.setError("code", { message: "This code is already in use" });
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, DISCOUNT_CODE_FORM_FIELDS);
          return;
        }
        toast.error(error.message);
      },
    });
  }

  const isPending = createDiscountCode.isPending || updateDiscountCode.isPending;
  const isPercent = discountType === DISCOUNT_TYPE.percent;

  return (
    <Form {...form}>
      <form
        className="dashboard-profile-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="code"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Code">
                <Input {...field} className="uppercase" />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="discount_type"
          render={({ field }) => (
            <FormItem>
              <FieldControl label="Discount type">
                <Select
                  onChange={(event) => {
                    const nextType = event.target
                      .value as DiscountCodeFormInput["discount_type"];
                    field.onChange(nextType);
                    if (nextType === DISCOUNT_TYPE.percent) {
                      const currentValue = form.getValues("value");
                      if (
                        currentValue !== undefined &&
                        currentValue > 100
                      ) {
                        form.setValue("value", 100, { shouldValidate: true });
                      }
                    }
                  }}
                  value={field.value}
                >
                  <option value={DISCOUNT_TYPE.percent}>Percent</option>
                  <option value={DISCOUNT_TYPE.fixedCents}>
                    Fixed amount (cents)
                  </option>
                </Select>
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="value"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={
                  isPercent
                    ? FORM_FIELD_HINT.discountPercentOff
                    : FORM_FIELD_HINT.catalogPriceCents
                }
                label={isPercent ? "Percent off" : "Discount amount (cents)"}
              >
                <IntegerFieldInput
                  allowEmpty
                  min={1}
                  max={isPercent ? 100 : undefined}
                  {...field}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="min_order_cents"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_FIELD_HINT.discountMinOrderCents}
                label="Minimum order (cents)"
                optional
              >
                <OptionalIntegerFieldInput
                  min={0}
                  {...field}
                  value={field.value ?? null}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="max_uses"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_FIELD_HINT.discountMaxUses}
                label="Maximum uses"
                optional
              >
                <OptionalIntegerFieldInput
                  min={1}
                  {...field}
                  value={field.value ?? null}
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

        <FormField
          control={form.control}
          name="is_active"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={FORM_SWITCH_HINT.checkoutActive}
                hintId="discount-checkout-active-hint"
                label={FORM_SWITCH_LABEL.checkoutActive}
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

        {mode === "edit" && discountCode ? (
          <p className="text-sm text-muted">
            Used {discountCode.used_count} time
            {discountCode.used_count === 1 ? "" : "s"}
          </p>
        ) : null}

        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel={mode === "create" ? "Create discount code" : "Save changes"}
            isPending={isPending}
            pendingLabel={mode === "create" ? "Creating…" : "Saving…"}
          />
        </div>
      </form>
    </Form>
  );
}
