import { z } from "zod";

import {
  CLOUDINARY_MAX_PRODUCT_IMAGES,
  isCloudinaryConfigured,
  isCloudinaryDeliveryUrlInFolder,
} from "@/lib/cloudinary/config";
import { CLOUDINARY_UPLOAD_COPY } from "@/lib/cloudinary/messages";

const productImageItemSchema = z
  .string()
  .trim()
  .max(2048)
  .url("Enter a valid URL")
  .refine(
    (value) => value.startsWith("http://") || value.startsWith("https://"),
    "URL must use HTTP or HTTPS",
  )
  .superRefine((value, ctx) => {
    if (!isCloudinaryConfigured()) {
      return;
    }
    if (!isCloudinaryDeliveryUrlInFolder(value)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: CLOUDINARY_UPLOAD_COPY.productImageValidation,
      });
    }
  });

const productImageUrlResponseItemSchema = z
  .string()
  .trim()
  .max(2048)
  .url()
  .refine(
    (value) => value.startsWith("http://") || value.startsWith("https://"),
    "URL must use HTTP or HTTPS",
  );

/** Lenient gallery URLs for API responses — backend already sanitizes on write. */
export const productImageUrlsResponseSchema = z
  .union([
    z
      .array(productImageUrlResponseItemSchema)
      .max(
        CLOUDINARY_MAX_PRODUCT_IMAGES,
        `Maximum ${CLOUDINARY_MAX_PRODUCT_IMAGES} images allowed`,
      ),
    z.null(),
    z.undefined(),
  ])
  .transform((urls) => urls ?? []);

/** Product gallery URLs for forms — Cloudinary folder rules apply when configured (max 5). */
export const productImageUrlsSchema = z
  .array(productImageItemSchema)
  .max(
    CLOUDINARY_MAX_PRODUCT_IMAGES,
    `Maximum ${CLOUDINARY_MAX_PRODUCT_IMAGES} images allowed`,
  )
  .refine(
    (urls) => new Set(urls).size === urls.length,
    "Duplicate image URLs are not allowed",
  );
