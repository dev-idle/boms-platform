import { z } from "zod";

import { isCloudinaryConfigured, isCloudinaryDeliveryUrlInFolder } from "@/lib/cloudinary/config";
import { optionalHttpUrlSchema } from "@/lib/validation/url";

/** Product image URL — Cloudinary product folder only when NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME is set. */
export const productImageUrlSchema = optionalHttpUrlSchema.superRefine(
  (value, ctx) => {
    if (!value || !isCloudinaryConfigured()) {
      return;
    }
    if (!isCloudinaryDeliveryUrlInFolder(value)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Upload a product image or use a URL from the configured Cloudinary folder",
      });
    }
  },
);
