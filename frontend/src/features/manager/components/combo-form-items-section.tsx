"use client";

import { useMemo, useId, useState } from "react";
import {
  useWatch,
  type Control,
  type UseFieldArrayAppend,
  type UseFieldArrayRemove,
  type FieldArrayWithId,
} from "react-hook-form";

import { Button } from "@/components/ui/button";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
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
import { CATALOG_INTEGER_MAX } from "@/lib/validation/catalog";

import type { ComboFormValues, ManagerCombo } from "../schemas";

import {
  ComboProductPicker,
  type ComboProductPickerValue,
} from "./combo-product-picker";

type ComboFormItemsSectionProps = {
  append: UseFieldArrayAppend<ComboFormValues, "items">;
  combo?: ManagerCombo;
  control: Control<ComboFormValues>;
  fields: FieldArrayWithId<ComboFormValues, "items", "id">[];
  remove: UseFieldArrayRemove;
};

export function ComboFormItemsSection({
  append,
  combo,
  control,
  fields,
  remove,
}: ComboFormItemsSectionProps) {
  const hintId = useId();
  const addQuantityId = useId();
  const [draftProduct, setDraftProduct] = useState<ComboProductPickerValue>(null);
  const [draftQuantity, setDraftQuantity] = useState(1);
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

  /**
   * A bundle of one product needs a quantity of at least two, which the hint
   * above says out loud — so the composer asks for the quantity where the
   * manager already is, instead of adding a row they must then go and edit.
   *
   * Add stays disabled until there is something to add: a button that cannot
   * act says so by being greyed, not by raising an error after the click.
   */
  function handleAddProduct(): void {
    if (!draftProduct || takenProductIds.has(draftProduct.id)) {
      return;
    }
    append(
      { product_id: draftProduct.id, quantity: draftQuantity },
      { shouldFocus: false },
    );
    setSessionProductNames((previous) => ({
      ...previous,
      [draftProduct.id]: draftProduct.name,
    }));
    setDraftProduct(null);
    setDraftQuantity(1);
  }

  return (
    <section
      aria-labelledby="combo-items-title"
      className="combo-form-items"
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
                          max={CATALOG_INTEGER_MAX}
                          min={1}
                          {...quantityField}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="combo-form-items__actions">
                  <DashboardTableActionButton
                    label={`Remove ${productName}`}
                    onClick={() => remove(index)}
                    text="Remove"
                    tone="danger"
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
          onChange={setDraftProduct}
          value={draftProduct}
        />
        <div className="combo-form-items__qty">
          <IntegerFieldInput
            aria-label="Quantity to add"
            id={addQuantityId}
            max={CATALOG_INTEGER_MAX}
            min={1}
            onChange={(value) => setDraftQuantity(value ?? 1)}
            value={draftQuantity}
          />
        </div>
        <div className="combo-form-items__actions">
          <Button
            aria-label="Add product to bundle"
            disabled={!draftProduct}
            // A greyed control says why it is greyed, here as everywhere else.
            title={draftProduct ? undefined : "Pick a product to add."}
            onClick={handleAddProduct}
            // Mousedown would blur the picker and close its list before the
            // click lands on a button that no longer exists.
            onMouseDown={(event) => {
              event.preventDefault();
            }}
            size="sm"
            type="button"
            variant="outline"
          >
            Add
          </Button>
        </div>
      </div>

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
