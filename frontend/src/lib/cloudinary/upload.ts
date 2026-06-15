import { z } from "zod";

import { getCloudinaryCloudName, isCloudinaryDeliveryUrlInFolder } from "./config";
import { CLOUDINARY_UPLOAD_COPY, cloudinaryImageTooLargeMessage } from "./messages";
import type { CloudinaryUploadSignature } from "./signature";

function cloudinaryUploadResultSchema(maxBytes: number) {
  return z.object({
    bytes: z.number().int().positive().max(maxBytes),
    secure_url: z.string().url(),
  });
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
  formData.append("max_file_size", String(signature.max_bytes));

  const response = await fetch(signature.upload_url, {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    throw new Error(CLOUDINARY_UPLOAD_COPY.uploadFailedRemote);
  }

  const payload: unknown = await response.json();
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
