import {
  CLOUDINARY_ALLOWED_IMAGE_TYPES,
  CLOUDINARY_MAX_IMAGE_BYTES,
  CLOUDINARY_MAX_PRODUCT_IMAGES,
} from "./config";
import { isDuplicateCatalogFileLabel, normalizeCatalogFileLabel } from "./format";
import {
  CLOUDINARY_UPLOAD_COPY,
  cloudinaryImageTooLargeMessage,
  cloudinaryTooManyProductImagesMessage,
} from "./messages";

type PlanCatalogImageUploadInput = {
  allowedTypes?: readonly string[];
  existingUrls: string[];
  files: File[];
  labelsByUrl: Record<string, string>;
  maxBytes?: number;
  maxImages?: number;
};

export function planCatalogImageUpload({
  allowedTypes = CLOUDINARY_ALLOWED_IMAGE_TYPES,
  existingUrls,
  files,
  labelsByUrl,
  maxBytes = CLOUDINARY_MAX_IMAGE_BYTES,
  maxImages = CLOUDINARY_MAX_PRODUCT_IMAGES,
}: PlanCatalogImageUploadInput):
  | { files: File[]; ok: true }
  | { message: string; ok: false } {
  if (files.length === 0) {
    return { message: "", ok: false };
  }

  const remainingSlots = maxImages - existingUrls.length;
  if (files.length > remainingSlots) {
    return {
      message: cloudinaryTooManyProductImagesMessage(
        existingUrls.length,
        files.length,
        maxImages,
      ),
      ok: false,
    };
  }

  const seenInBatch = new Set<string>();

  for (const file of files) {
    if (!allowedTypes.includes(file.type as (typeof allowedTypes)[number])) {
      return { message: CLOUDINARY_UPLOAD_COPY.allowedFormats, ok: false };
    }
    if (file.size > maxBytes) {
      return { message: cloudinaryImageTooLargeMessage(maxBytes), ok: false };
    }

    const normalized = normalizeCatalogFileLabel(file.name);
    if (normalized) {
      if (seenInBatch.has(normalized)) {
        return { message: CLOUDINARY_UPLOAD_COPY.duplicateFileName, ok: false };
      }
      if (isDuplicateCatalogFileLabel(file.name, existingUrls, labelsByUrl)) {
        return { message: CLOUDINARY_UPLOAD_COPY.duplicateFileName, ok: false };
      }
      seenInBatch.add(normalized);
    }
  }

  return { files, ok: true };
}

export function cloudinaryUploadSuccessMessage(count: number): string {
  return count === 1
    ? CLOUDINARY_UPLOAD_COPY.uploadSuccess
    : CLOUDINARY_UPLOAD_COPY.uploadSuccessCount(count);
}
