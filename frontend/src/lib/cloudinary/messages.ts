import { CLOUDINARY_MAX_IMAGE_BYTES, CLOUDINARY_MAX_PRODUCT_IMAGES } from "./config";
import { formatCloudinaryMaxImageSize } from "./format";

/** User-facing Cloudinary upload copy (SSOT). */
export const CLOUDINARY_UPLOAD_COPY = {
  allowedFormats: "Use JPG, PNG, WebP, or AVIF",
  cloudNameMismatch:
    "Cloudinary cloud name does not match this environment. Check NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME.",
  folderMismatch:
    "Cloudinary folder does not match this environment. Check NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER.",
  duplicateImage: "This image is already in the list",
  productImageValidation:
    "Upload a product image or use a URL from the configured Cloudinary folder",
  uploadFailed: "Failed to upload image",
  uploadFailedRemote: "Image upload failed",
  uploadInvalidResponse: "Image upload returned an invalid response",
  uploadSuccess: "Image uploaded",
  uploadUrlNotAllowed: "Uploaded image URL is not allowed",
} as const;

export function cloudinaryImageTooLargeMessage(maxBytes: number): string {
  return `Image must be ${formatCloudinaryMaxImageSize(maxBytes)} or smaller`;
}

export function cloudinaryProductImageFieldHint(
  maxBytes = CLOUDINARY_MAX_IMAGE_BYTES,
  maxImages = CLOUDINARY_MAX_PRODUCT_IMAGES,
): string {
  return `JPG, PNG, WebP, or AVIF up to ${formatCloudinaryMaxImageSize(maxBytes)} each (max ${maxImages} images).`;
}
