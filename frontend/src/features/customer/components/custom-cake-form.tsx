"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState, type ReactNode } from "react";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { CLOUDINARY_MAX_IMAGE_BYTES } from "@/lib/cloudinary/config";
import { formatCloudinaryMaxImageSize } from "@/lib/cloudinary/format";
import { CLOUDINARY_UPLOAD_COPY } from "@/lib/cloudinary/messages";
import { isApiError, userFacingErrorMessage } from "@/lib/errors";
import {
  CAKE_MESSAGE_MAX,
  PRODUCT_OPTION_GROUP_LABEL,
  productOptionGroupSchema,
  type CatalogProduct,
} from "@/lib/schemas/catalog";
import { cn } from "@/lib/utils";
import { applyFormFieldErrors } from "@/lib/validation";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useAddCartItem } from "../hooks";
import { cartCustomizationInputSchema, type CartCustomizationInput } from "../schemas";
import { ReferencePhotoField } from "./reference-photo-field";

const FIELDS = ["option_ids", "message", "reference_image_url", "reference_rights_confirmed"] as const;

type CustomCakeFormProps = {
  product: CatalogProduct;
  /** What stands beside the submit. */
  children?: ReactNode;
  /** Told what the chosen options add to the product's price, as they change. */
  onOptionsCentsChange?: (cents: number) => void;
};

/**
 * A custom cake configured before it goes in the cart: one option from each
 * group it offers, a message for the cake, and a reference photo the customer
 * uploads and says they may share. The page's price follows the choices.
 */
export function CustomCakeForm({ product, children, onOptionsCentsChange }: CustomCakeFormProps) {
  const groups = productOptionGroupSchema.options.filter((group) =>
    product.options.some((option) => option.group === group),
  );
  const empty: CartCustomizationInput = {
    option_ids: groups.map(() => ""),
    message: "",
    reference_image_url: "",
    reference_rights_confirmed: false,
  };
  const form = useForm<CartCustomizationInput>({
    resolver: zodResolver(cartCustomizationInputSchema),
    defaultValues: empty,
  });
  const addCartItem = useAddCartItem();
  const [uploading, setUploading] = useState(false);
  const photo = useWatch({ control: form.control, name: "reference_image_url" });
  const busy = uploading || addCartItem.isPending;

  /** What the chosen options add to the price once optionId is the choice of group index. */
  function optionsCents(index: number, optionId: string): number {
    const chosen = form.getValues("option_ids").map((id, at) => (at === index ? optionId : id));
    return product.options
      .filter((option) => chosen.includes(option.id))
      .reduce((total, option) => total + option.price_delta_cents, 0);
  }

  function onSubmit(values: CartCustomizationInput): void {
    addCartItem.mutate(
      { product_id: product.id, quantity: 1, customization: values },
      {
        onSuccess: () => {
          form.reset(empty);
          onOptionsCentsChange?.(0);
        },
        onError: (error) => {
          if (isApiError(error) && error.hasValidationDetails()) {
            applyFormFieldErrors(form, error.details!, FIELDS);
            return;
          }
          toast.error(userFacingErrorMessage(error, "We could not add it to your cart. Please try again."));
        },
      },
    );
  }

  return (
    <Form {...form}>
      <form
        className="dashboard-profile-form storefront-account-form"
        method="post"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        {groups.map((group, index) => (
          <FormField
            key={group}
            control={form.control}
            name={`option_ids.${index}`}
            render={({ field }) => (
              <FormItem>
                {/* One tile per option, so every choice and its price shows at once;
                    FormControl ties the group to its error message. */}
                <FormControl>
                  <fieldset className="option-tiles" disabled={busy}>
                    <legend className="option-tiles__legend text-form-label">{PRODUCT_OPTION_GROUP_LABEL[group]}</legend>
                    <div className="option-tiles__grid">
                      {product.options
                        .filter((option) => option.group === group)
                        .map((option) => (
                          <label
                            className={cn("option-tile", field.value === option.id && "option-tile--selected")}
                            key={option.id}
                          >
                            <input
                              checked={field.value === option.id}
                              className="sr-only"
                              name={field.name}
                              onBlur={field.onBlur}
                              onChange={() => {
                                field.onChange(option.id);
                                onOptionsCentsChange?.(optionsCents(index, option.id));
                              }}
                              type="radio"
                              value={option.id}
                            />
                            <span className="option-tile__label">{option.label}</span>
                            <span className="option-tile__note">
                              {option.price_delta_cents > 0 ? `+${formatPriceCents(option.price_delta_cents)}` : "Included"}
                            </span>
                          </label>
                        ))}
                    </div>
                  </fieldset>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        ))}
        <FormField
          control={form.control}
          name="message"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`Written on the cake, up to ${CAKE_MESSAGE_MAX} characters.`}
                hintId="cake-message-hint"
                label="Message"
                optional
              >
                <Input disabled={busy} maxLength={CAKE_MESSAGE_MAX} placeholder="Happy birthday, Mai" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="reference_image_url"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`A photo of a cake you like, ${CLOUDINARY_UPLOAD_COPY.allowedFormats} up to ${formatCloudinaryMaxImageSize(CLOUDINARY_MAX_IMAGE_BYTES)}. Our bakers use it as a guide.`}
                hintId="cake-photo-hint"
                label="Reference photo"
                optional
              >
                <ReferencePhotoField
                  disabled={busy}
                  photo={field.value}
                  uploading={uploading}
                  onChange={(url) => {
                    field.onChange(url);
                    if (url === "") {
                      form.setValue("reference_rights_confirmed", false);
                    }
                  }}
                  onUploadingChange={setUploading}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {photo ? (
          <FormField
            control={form.control}
            name="reference_rights_confirmed"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <CheckboxField
                    checked={field.value}
                    disabled={busy}
                    onChange={(event) => field.onChange(event.target.checked)}
                  >
                    I took this photo or may share it, and the bakery may use it to make my cake.
                  </CheckboxField>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        ) : null}
        <p className="text-caption">
          Without a reference photo, your cake is confirmed when you pay. With one, we check it first and accept or
          decline it within 24 hours of your payment. If we cannot make it, or have not accepted it by then, you are
          refunded in full.
        </p>
        <div className="dashboard-profile-form-actions">
          <Button aria-busy={addCartItem.isPending || undefined} disabled={busy} type="submit">
            {addCartItem.isPending ? "Adding…" : "Add to cart"}
          </Button>
          {children}
        </div>
      </form>
    </Form>
  );
}
