import { CLOUDINARY_MAX_IMAGE_BYTES, CLOUDINARY_MAX_PRODUCT_IMAGES } from "./config";
import { formatCloudinaryMaxImageSize } from "./format";

/** User-facing Cloudinary upload copy (SSOT). */
export const CLOUDINARY_UPLOAD_COPY = {
  allowedFormats: "JPG, PNG, WebP, or AVIF",
  cloudNameMismatch:
    "Cloudinary cloud name does not match this environment. Check NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME.",
  folderMismatch:
    "Cloudinary folder does not match this environment. Check NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER.",
  duplicateImage: "This image is already in the list",
  duplicateFileName: "A file with this name is already in the list",
  productImageValidation:
    "Upload a product image or use a URL from the configured Cloudinary folder",
  uploadFailed: "Failed to upload image",
  uploadFailedRemote: "Image upload failed",
  uploadServiceUnavailable:
    "Image upload is not available. Configure Cloudinary on the API (CLOUDINARY_* env vars).",
  uploadInvalidResponse: "Image upload returned an invalid response",
  uploadSuccess: "Image uploaded",
  uploadSuccessCount: (count: number) => `${count} images uploaded`,
  uploadUrlNotAllowed: "Uploaded image URL is not allowed",
} as const;

export function cloudinaryTooManyProductImagesMessage(
  currentCount: number,
  selectedCount: number,
  maxImages: number,
): string {
  if (currentCount >= maxImages) {
    return `This product already has the maximum of ${maxImages} images`;
  }

  const remainingSlots = maxImages - currentCount;
  const slotLabel = remainingSlots === 1 ? "slot" : "slots";
  return `You selected ${selectedCount} images but only ${remainingSlots} ${slotLabel} remain (${currentCount} of ${maxImages})`;
}

export function cloudinaryImageTooLargeMessage(maxBytes: number): string {
  return `Image must be ${formatCloudinaryMaxImageSize(maxBytes)} or smaller`;
}

/** Field hint — upload limits (primary/order is shown in the gallery UI). */
export function cloudinaryProductImageFieldHint(
  maxBytes = CLOUDINARY_MAX_IMAGE_BYTES,
  maxImages = CLOUDINARY_MAX_PRODUCT_IMAGES,
): string {
  const sizeHint = formatCloudinaryMaxImageSize(maxBytes);
  return `Maximum ${maxImages} images. ${CLOUDINARY_UPLOAD_COPY.allowedFormats}, ${sizeHint} each.`;
}
