"use client";

import { useMemo, useId, useRef, useState } from "react";
import {
  useWatch,
  type Control,
  type UseFieldArrayAppend,
  type UseFieldArrayRemove,
  type FieldArrayWithId,
} from "react-hook-form";

import { DashboardTableAddButton, DashboardTableDeleteButton } from "@/components/ui/dashboard-table-actions";
import { FormFieldHint } from "@/components/ui/form-field-hint";
import { FORM_FIELD_HINT } from "@/constants/dashboard-form-copy";
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { IntegerFieldInput } from "@/components/ui/integer-field-input";
import { Label } from "@/components/ui/label";

import type { ComboFormInput, ManagerCombo } from "../schemas";

import {
  ComboProductPicker,
  type ComboProductPickerValue,
} from "./combo-product-picker";

type ComboFormItemsSectionProps = {
  append: UseFieldArrayAppend<ComboFormInput, "items">;
  combo?: ManagerCombo;
  control: Control<ComboFormInput>;
  fields: FieldArrayWithId<ComboFormInput, "items", "id">[];
  remove: UseFieldArrayRemove;
};

export function ComboFormItemsSection({
  append,
  combo,
  control,
  fields,
  remove,
}: ComboFormItemsSectionProps) {
  const addQuantityId = useId();
  const hintId = useId();
  const sectionRef = useRef<HTMLElement>(null);
  const addQuantityRef = useRef<HTMLInputElement>(null);
  const [draftProduct, setDraftProduct] = useState<ComboProductPickerValue>(null);
  const [draftQuantity, setDraftQuantity] = useState(1);
  const [addError, setAddError] = useState<string | null>(null);

  const [sessionProductNames, setSessionProductNames] = useState<
    Record<string, string>
  >(() =>
    Object.fromEntries(
      (combo?.items ?? []).map((item) => [item.product_id, item.product_name]),
    ),
  );

  const watchedItems = useWatch({ control, name: "items" });

  const takenProductIds = useMemo(
    () =>
      new Set(
        (watchedItems ?? [])
          .map((item) => item.product_id)
          .filter((id): id is string => Boolean(id)),
      ),
    [watchedItems],
  );

  const productNameById = useMemo(
    () => new Map(Object.entries(sessionProductNames)),
    [sessionProductNames],
  );

  function releaseComposerFocus(): void {
    addQuantityRef.current?.blur();
    const active = document.activeElement;
    if (
      active instanceof HTMLElement &&
      sectionRef.current?.contains(active) &&
      active.closest(".combo-form-items__qty")
    ) {
      active.blur();
    }
  }

  function handleAddProduct(): void {
    if (!draftProduct) {
      setAddError("Pick a product from the list");
      return;
    }
    if (takenProductIds.has(draftProduct.id)) {
      setAddError("This product is already in the bundle");
      return;
    }

    append(
      { product_id: draftProduct.id, quantity: draftQuantity },
      { shouldFocus: false },
    );
    setSessionProductNames((prev) => ({
      ...prev,
      [draftProduct.id]: draftProduct.name,
    }));
    setDraftProduct(null);
    setDraftQuantity(1);
    setAddError(null);
    queueMicrotask(releaseComposerFocus);
  }

  return (
    <section
      aria-labelledby="combo-items-title"
      className="combo-form-items"
      ref={sectionRef}
    >
      <div className="field-control-label-block">
        <Label id="combo-items-title">Bundle contents</Label>
        <FormFieldHint id={hintId}>{FORM_FIELD_HINT.comboItems}</FormFieldHint>
      </div>

      {fields.length > 0 ? (
        <ul className="combo-form-items__list">
          {fields.map((field, index) => {
            const productId = watchedItems?.[index]?.product_id ?? "";
            const productName =
              productNameById.get(productId) ?? "Unknown product";

            return (
              <li key={field.id} className="combo-form-items__line">
                <span className="combo-form-items__name">{productName}</span>
                <FormField
                  control={control}
                  name={`items.${index}.quantity`}
                  render={({ field: quantityField }) => (
                    <FormItem className="combo-form-items__qty">
                      <FormControl>
                        <IntegerFieldInput
                          aria-label={`Quantity for ${productName}`}
                          min={1}
                          {...quantityField}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="combo-form-items__actions">
                  <DashboardTableDeleteButton
                    label={`Remove ${productName}`}
                    onClick={() => remove(index)}
                  />
                </div>
              </li>
            );
          })}
        </ul>
      ) : null}

      <div aria-describedby={hintId} className="combo-form-items__composer">
        <ComboProductPicker
          excludedProductIds={takenProductIds}
          onChange={(next) => {
            setDraftProduct(next);
            setAddError(null);
          }}
          value={draftProduct}
        />
        <div className="combo-form-items__qty">
          <IntegerFieldInput
            aria-label="Quantity to add"
            id={addQuantityId}
            min={1}
            onChange={(value) => setDraftQuantity(value ?? 1)}
            ref={addQuantityRef}
            value={draftQuantity}
          />
        </div>
        <div className="combo-form-items__actions">
          <DashboardTableAddButton
            label="Add product to bundle"
            onClick={handleAddProduct}
            onMouseDown={(event) => {
              event.preventDefault();
            }}
          />
        </div>
      </div>

      {addError ? (
        <p className="combo-form-items__error" role="alert">
          {addError}
        </p>
      ) : null}

      <FormField
        control={control}
        name="items"
        render={() => (
          <FormItem>
            <FormMessage />
          </FormItem>
        )}
      />
    </section>
  );
}
