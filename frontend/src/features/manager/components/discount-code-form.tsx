"use client";

import { useMemo } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FormPublishToggle } from "@/components/ui/form-publish-toggle";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { FieldControl } from "@/components/ui/field-control";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";
import {
  fromDatetimeLocalValue,
  toDatetimeLocalValue,
} from "@/lib/validation/datetime";

import {
  useCreateDiscountCode,
  useUpdateDiscountCode,
} from "../hooks";
import {
  DISCOUNT_TYPE,
  discountCodeFormSchema,
  type DiscountCodeFormInput,
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
    (): DiscountCodeFormInput => ({
      code: discountCode?.code ?? "",
      discount_type: discountCode?.discount_type ?? DISCOUNT_TYPE.percent,
      value: discountCode?.value ?? 10,
      min_order_cents: discountCode?.min_order_cents ?? null,
      max_uses: discountCode?.max_uses ?? null,
      starts_at: discountCode?.starts_at ?? defaultStartsAt(),
      ends_at: discountCode?.ends_at ?? defaultEndsAt(),
      is_active: discountCode?.is_active ?? true,
    }),
    [discountCode],
  );

  const form = useForm<DiscountCodeFormInput>({
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
                <Select onChange={field.onChange} value={field.value}>
                  <option value={DISCOUNT_TYPE.percent}>Percent</option>
                  <option value={DISCOUNT_TYPE.fixedCents}>Fixed amount (cents)</option>
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
                label={
                  discountType === DISCOUNT_TYPE.percent
                    ? "Percent off (1–100)"
                    : "Discount amount (cents)"
                }
              >
                <Input min={1} step={1} type="number" {...field} />
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
              <FieldControl label="Minimum order (cents)" optional>
                <Input
                  min={0}
                  step={1}
                  type="number"
                  value={field.value ?? ""}
                  onChange={(event) => {
                    const raw = event.target.value;
                    field.onChange(raw === "" ? null : Number(raw));
                  }}
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
              <FieldControl label="Maximum uses" optional>
                <Input
                  min={1}
                  step={1}
                  type="number"
                  value={field.value ?? ""}
                  onChange={(event) => {
                    const raw = event.target.value;
                    field.onChange(raw === "" ? null : Number(raw));
                  }}
                />
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
            <FormItem>
              <FormControl>
                <FormPublishToggle
                  checked={field.value}
                  description="When on, customers can apply this code at checkout."
                  id="discount-code-is-active"
                  label="Active"
                  onCheckedChange={field.onChange}
                />
              </FormControl>
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
