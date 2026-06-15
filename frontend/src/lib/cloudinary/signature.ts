import { z } from "zod";

import { browserRequest } from "@/lib/browser-api-client";

import {
  getCloudinaryCloudName,
  getCloudinaryUploadFolder,
} from "./config";
import { CLOUDINARY_UPLOAD_COPY } from "./messages";

export const cloudinaryUploadSignatureSchema = z.object({
  allowed_formats: z.string().min(1),
  api_key: z.string().min(1),
  cloud_name: z.string().min(1),
  folder: z.string().min(1),
  max_bytes: z.number().int().positive(),
  signature: z.string().min(1),
  timestamp: z.number().int().positive(),
  unique_filename: z.string().min(1),
  upload_url: z.string().url(),
});

export type CloudinaryUploadSignature = z.infer<
  typeof cloudinaryUploadSignatureSchema
>;

/** Ensures API signature targets match NEXT_PUBLIC_CLOUDINARY_* (fail-fast on env drift). */
export function validateCloudinarySignatureEnv(
  signature: CloudinaryUploadSignature,
): void {
  const cloudName = getCloudinaryCloudName();
  if (cloudName && signature.cloud_name !== cloudName) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.cloudNameMismatch);
  }

  const folder = getCloudinaryUploadFolder();
  if (signature.folder !== folder) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.folderMismatch);
  }
}

export async function fetchCloudinaryUploadSignature(): Promise<CloudinaryUploadSignature> {
  const signature = await browserRequest<CloudinaryUploadSignature>(
    "/api/v1/manager/media/cloudinary-signature",
    {
      method: "GET",
      schema: cloudinaryUploadSignatureSchema,
    },
  );
  validateCloudinarySignatureEnv(signature);
  return signature;
}
