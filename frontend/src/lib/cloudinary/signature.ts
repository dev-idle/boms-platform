import { z } from "zod";

import { browserRequest } from "@/lib/browser-api-client";

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

export function fetchCloudinaryUploadSignature(): Promise<CloudinaryUploadSignature> {
  return browserRequest<CloudinaryUploadSignature>(
    "/api/v1/manager/media/cloudinary-signature",
    {
      method: "GET",
      schema: cloudinaryUploadSignatureSchema,
    },
  );
}
