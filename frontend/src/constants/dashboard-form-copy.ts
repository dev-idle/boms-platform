/**
 * Dashboard form copy — hints and switch labels (SSOT).
 *
 * Hints (`FORM_FIELD_HINT`): text fields only — between label and control.
 *
 * Switch hints (`FORM_SWITCH_HINT`): under the toggle label, left column.
 * State the off-state effect; do not paraphrase the label.
 */

export const FORM_FIELD_HINT = {
  catalogPriceCents: "Amount in cents (100 = $1.00).",
  catalogProductImages:
    "First image is the storefront cover; gallery order matches what customers see.",
  catalogSlugCreate: "Auto-filled from name.",
  catalogSortOrder: "Lower values appear first.",
  operationalEmployeeCode: "Permanent staff ID. Assigned at creation.",
  productImageUrlFallback:
    "First image is the storefront cover; gallery order matches what customers see. Add up to 5 HTTPS URLs when Cloudinary is not configured.",
} as const;

export const FORM_SWITCH_LABEL = {
  storefrontVisible: "Visible on storefront",
  availableToOrder: "Available to order",
  checkoutActive: "Active at checkout",
} as const;

export const FORM_SWITCH_HINT = {
  storefrontVisible: "When off, hidden from browse and category filters.",
  availableToOrder: "When off, hidden from the storefront and cart.",
  checkoutActive: "When off, cannot be applied at checkout.",
} as const;

/** Catalog image upload — control copy; format/size note from `cloudinaryProductImageUploadNote()`. */
export const CATALOG_IMAGE_FIELD_COPY = {
  actionsAriaLabel: "Product image actions",
  addImage: "Add images",
  addImageUrl: "Add image URL",
  chooseFile: "Choose file",
  imageCount: (current: number, max: number) => `${current} of ${max} images`,
  imageListAriaLabel: "Product images",
  imagePosition: (position: number, total: number) => `Gallery image ${position} of ${total}`,
  primaryImage: "Primary",
  setPrimary: "Set as primary",
  setPrimaryImage: (name: string) => `Set ${name} as primary image`,
  noFile: "No file chosen",
  noImages: "No images added yet",
  uploadAriaLabel: "Choose product image files",
  uploadProgress: "Uploading images…",
  uploading: "Uploading…",
  replace: "Replace",
  remove: "Remove",
  view: "View",
  viewImage: (name: string) => `View ${name}`,
  viewImageDialogTitle: "Image preview",
  closePreview: "Close",
  previewLoadError: "Failed to load image preview.",
  previewLoading: "Loading image…",
} as const;
