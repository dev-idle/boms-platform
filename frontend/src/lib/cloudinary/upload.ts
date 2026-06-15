import { z } from "zod";

import { getCloudinaryCloudName, isCloudinaryDeliveryUrlInFolder } from "./config";
import { CLOUDINARY_UPLOAD_COPY, cloudinaryImageTooLargeMessage } from "./messages";
import type { CloudinaryUploadSignature } from "./signature";

const cloudinaryUploadErrorSchema = z.object({
  error: z.object({
    message: z.string().min(1),
  }),
});

function cloudinaryUploadResultSchema(maxBytes: number) {
  return z
    .object({
      bytes: z.coerce.number().int().positive().optional(),
      secure_url: z.string().url(),
    })
    .superRefine((value, ctx) => {
      if (value.bytes != null && value.bytes > maxBytes) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: cloudinaryImageTooLargeMessage(maxBytes),
        });
      }
    });
}

function cloudinaryUploadErrorMessage(payload: unknown): string | undefined {
  const parsed = cloudinaryUploadErrorSchema.safeParse(payload);
  return parsed.success ? parsed.data.error.message : undefined;
}

export async function uploadImageToCloudinary(
  file: File,
  signature: CloudinaryUploadSignature,
): Promise<string> {
  if (file.size > signature.max_bytes) {
    throw new Error(cloudinaryImageTooLargeMessage(signature.max_bytes));
  }

  const formData = new FormData();
  formData.append("file", file);
  formData.append("api_key", signature.api_key);
  formData.append("timestamp", String(signature.timestamp));
  formData.append("signature", signature.signature);
  formData.append("folder", signature.folder);
  formData.append("allowed_formats", signature.allowed_formats);
  formData.append("unique_filename", signature.unique_filename);

  const response = await fetch(signature.upload_url, {
    method: "POST",
    body: formData,
  });

  const payload: unknown = await response.json().catch(() => null);

  if (!response.ok) {
    throw new Error(
      cloudinaryUploadErrorMessage(payload) ?? CLOUDINARY_UPLOAD_COPY.uploadFailedRemote,
    );
  }

  const parsed = cloudinaryUploadResultSchema(signature.max_bytes).safeParse(payload);
  if (!parsed.success) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.uploadInvalidResponse);
  }

  const cloudName = getCloudinaryCloudName();
  if (
    cloudName &&
    !isCloudinaryDeliveryUrlInFolder(parsed.data.secure_url, signature.folder, cloudName)
  ) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.uploadUrlNotAllowed);
  }

  return parsed.data.secure_url;
}
