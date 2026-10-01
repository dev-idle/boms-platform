"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CheckboxField } from "@/components/ui/checkbox-field";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { CLOUDINARY_MAX_IMAGE_BYTES } from "@/lib/cloudinary/config";
import { formatCloudinaryMaxImageSize } from "@/lib/cloudinary/format";
import { CLOUDINARY_UPLOAD_COPY } from "@/lib/cloudinary/messages";
import { isApiError } from "@/lib/errors";
import {
  CAKE_MESSAGE_MAX,
  PRODUCT_OPTION_GROUP_LABEL,
  productOptionGroupSchema,
  type CatalogProduct,
} from "@/lib/schemas/catalog";
import { applyFormFieldErrors } from "@/lib/validation";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useAddCartItem } from "../hooks";
import { cartCustomizationInputSchema, type CartCustomizationInput } from "../schemas";
import { ReferencePhotoField } from "./reference-photo-field";

const FIELDS = ["option_ids", "message", "reference_image_url", "reference_rights_confirmed"] as const;

type CustomCakeFormProps = {
  product: CatalogProduct;
};

/**
 * A custom cake configured before it goes in the cart: one option from each
 * group it offers, a message for the cake, and a reference photo the customer
 * uploads and says they may share. The price follows the choices.
 */
export function CustomCakeForm({ product }: CustomCakeFormProps) {
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
  const chosen = useWatch({ control: form.control, name: "option_ids" });
  const photo = useWatch({ control: form.control, name: "reference_image_url" });
  const price = product.options
    .filter((option) => chosen.includes(option.id))
    .reduce((total, option) => total + option.price_delta_cents, product.price_cents);
  const busy = uploading || addCartItem.isPending;

  function onSubmit(values: CartCustomizationInput): void {
    addCartItem.mutate(
      { product_id: product.id, quantity: 1, customization: values },
      {
        onSuccess: () => form.reset(empty),
        onError: (error) => {
          if (isApiError(error) && error.hasValidationDetails()) {
            applyFormFieldErrors(form, error.details!, FIELDS);
            return;
          }
          toast.error(isApiError(error) ? error.message : "Failed to add to cart");
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
                <FieldControl label={PRODUCT_OPTION_GROUP_LABEL[group]}>
                  <Select disabled={busy} value={field.value} onChange={field.onChange}>
                    <option value="">{`Choose a ${PRODUCT_OPTION_GROUP_LABEL[group].toLowerCase()}`}</option>
                    {product.options
                      .filter((option) => option.group === group)
                      .map((option) => (
                        <option key={option.id} value={option.id}>
                          {option.price_delta_cents > 0
                            ? `${option.label} (+${formatPriceCents(option.price_delta_cents)})`
                            : option.label}
                        </option>
                      ))}
                  </Select>
                </FieldControl>
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
          We confirm every custom cake before we start. If we cannot make it, you are refunded in full.
        </p>
        <div className="dashboard-profile-form-actions">
          <Button aria-busy={addCartItem.isPending || undefined} disabled={busy} type="submit">
            {addCartItem.isPending ? "Adding…" : `Add to cart · ${formatPriceCents(price)}`}
          </Button>
        </div>
      </form>
    </Form>
  );
}
