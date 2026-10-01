"use client";

import { useId } from "react";
import { useFieldArray, type Control } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { FormFieldHint } from "@/components/ui/form-field-hint";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { MoneyFieldInput } from "@/components/ui/money-field-input";
import { Select } from "@/components/ui/select";
import { FORM_FIELD_HINT } from "@/constants/dashboard-form-copy";
import { PRODUCT_OPTION_GROUP_LABEL, productOptionGroupSchema } from "@/lib/schemas/catalog";

import { MAX_PRODUCT_OPTIONS, type ProductFormValues } from "../schemas";

type ProductOptionsSectionProps = {
  control: Control<ProductFormValues>;
  disabled: boolean;
};

/** What a customer picks from on a customizable product: one option from each group offered. */
export function ProductOptionsSection({ control, disabled }: ProductOptionsSectionProps) {
  const hintId = useId();
  const { append, fields, remove } = useFieldArray({ control, name: "options", keyName: "key" });
  const full = fields.length >= MAX_PRODUCT_OPTIONS;

  return (
    <section aria-describedby={hintId} aria-labelledby="product-options-title" className="product-form-options">
      <div className="field-control-label-block">
        <Label id="product-options-title">Customer options</Label>
        <FormFieldHint id={hintId}>{FORM_FIELD_HINT.productOptions}</FormFieldHint>
      </div>

      {fields.length > 0 ? (
        <ul className="product-form-options__list">
          {fields.map((field, index) => (
            <li key={field.key} className="product-form-options__line">
              <FormField
                control={control}
                name={`options.${index}.group`}
                render={({ field: group }) => (
                  <FormItem>
                    <FormControl>
                      <Select
                        aria-label={`Group of option ${index + 1}`}
                        disabled={disabled}
                        value={group.value}
                        onChange={group.onChange}
                      >
                        {productOptionGroupSchema.options.map((value) => (
                          <option key={value} value={value}>
                            {PRODUCT_OPTION_GROUP_LABEL[value]}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={control}
                name={`options.${index}.label`}
                render={({ field: label }) => (
                  <FormItem>
                    <FormControl>
                      <Input
                        aria-label={`Name of option ${index + 1}`}
                        disabled={disabled}
                        maxLength={60}
                        placeholder="20 cm, serves 8–10"
                        {...label}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={control}
                name={`options.${index}.price_delta_cents`}
                render={({ field: price }) => (
                  <FormItem>
                    <FormControl>
                      <MoneyFieldInput aria-label={`Added price of option ${index + 1}`} disabled={disabled} {...price} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={control}
                name={`options.${index}.is_active`}
                render={({ field: active }) => (
                  <FormItem>
                    <CheckboxField
                      checked={active.value}
                      disabled={disabled}
                      onChange={(event) => active.onChange(event.target.checked)}
                    >
                      Offered
                    </CheckboxField>
                  </FormItem>
                )}
              />
              <div className="product-form-options__actions">
                <DashboardTableActionButton
                  blockedReason={disabled ? "Saving…" : undefined}
                  label={`Remove option ${index + 1}`}
                  onClick={() => remove(index)}
                  text="Remove"
                  tone="danger"
                />
              </div>
            </li>
          ))}
        </ul>
      ) : null}

      <div>
        <Button
          disabled={disabled || full}
          size="sm"
          title={full ? `A product offers at most ${MAX_PRODUCT_OPTIONS} options.` : undefined}
          type="button"
          variant="outline"
          onClick={() => append({ group: "size", label: "", price_delta_cents: 0, is_active: true })}
        >
          Add option
        </Button>
      </div>

      <FormField
        control={control}
        name="options"
        render={() => (
          <FormItem>
            <FormMessage />
          </FormItem>
        )}
      />
    </section>
  );
}
