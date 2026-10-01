import { z } from "zod";

import { browserRequest } from "@/lib/browser-api-client";

import {
  getCloudinaryCloudName,
  getCloudinaryUploadFolder,
} from "./config";
import { CLOUDINARY_UPLOAD_COPY } from "./messages";

const CLOUDINARY_UPLOAD_API_HOST = "api.cloudinary.com";

const cloudinaryUploadSignatureSchema = z.object({
  api_key: z.string().min(1),
  cloud_name: z.string().min(1),
  /** Where the image lands; the URL Cloudinary returns must be in it. */
  folder: z.string().min(1),
  max_bytes: z.number().int().positive(),
  /** The signed upload fields, sent with the file exactly as given. */
  params: z.record(z.string(), z.string()),
  signature: z.string().min(1),
  upload_url: z.url(),
});

export type CloudinaryUploadSignature = z.infer<
  typeof cloudinaryUploadSignatureSchema
>;

/**
 * Ensures API signature targets match NEXT_PUBLIC_CLOUDINARY_* (fail-fast on env
 * drift). `folder` is the folder this environment expects, when the browser
 * knows it: a customer's own reference folder is the API's to name.
 */
export function validateCloudinarySignatureEnv(
  signature: CloudinaryUploadSignature,
  folder?: string,
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

  if (folder !== undefined && signature.folder !== folder) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.folderMismatch);
  }
}

/** Signs a manager's catalog image upload. */
export async function fetchCloudinaryUploadSignature(): Promise<CloudinaryUploadSignature> {
  const signature = await browserRequest<CloudinaryUploadSignature>(
    "/api/v1/manager/media/cloudinary-signature",
    {
      method: "GET",
      schema: cloudinaryUploadSignatureSchema,
    },
  );
  validateCloudinarySignatureEnv(signature, getCloudinaryUploadFolder());
  return signature;
}

/** Signs a customer's reference photo upload into a folder of their own. */
export async function fetchReferenceUploadSignature(): Promise<CloudinaryUploadSignature> {
  const signature = await browserRequest<CloudinaryUploadSignature>("/api/v1/cart/reference-upload", {
    method: "GET",
    schema: cloudinaryUploadSignatureSchema,
  });
  validateCloudinarySignatureEnv(signature);
  return signature;
}
