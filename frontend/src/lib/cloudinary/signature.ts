import { z } from "zod";

import { browserRequest } from "@/lib/browser-api-client";

import {
  getCloudinaryCloudName,
  getCloudinaryUploadFolder,
} from "./config";
import { CLOUDINARY_UPLOAD_COPY } from "./messages";

const CLOUDINARY_UPLOAD_API_HOST = "api.cloudinary.com";

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
  let parsedUploadUrl: URL;
  try {
    parsedUploadUrl = new URL(signature.upload_url);
  } catch {
    throw new Error(CLOUDINARY_UPLOAD_COPY.uploadFailedRemote);
  }

  const expectedPath = `/v1_1/${signature.cloud_name}/image/upload`;
  if (
    parsedUploadUrl.protocol !== "https:" ||
    parsedUploadUrl.hostname !== CLOUDINARY_UPLOAD_API_HOST ||
    parsedUploadUrl.pathname !== expectedPath
  ) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.uploadFailedRemote);
  }

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
