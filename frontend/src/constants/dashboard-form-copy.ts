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
  catalogSlugCreate: "Auto-filled from name.",
  catalogSortOrder: "Lower values appear first.",
  discountMaxUses: "Leave empty for unlimited uses.",
  discountMinOrderCents: "Minimum cart total in cents. Leave empty for no minimum.",
  discountPercentOff: "Whole number from 1 to 100.",
  operationalEmployeeCode: "Permanent staff ID. Assigned at creation.",
  productImageUrlFallback: "Maximum 5 HTTPS image URLs.",
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

/** Catalog image upload — control copy; field hint from `cloudinaryProductImageFieldHint()`. */
export const CATALOG_IMAGE_FIELD_COPY = {
  actionsAriaLabel: "Product image actions",
  addImage: "Add images",
  addImageUrl: "Add image URL",
  chooseFile: "Choose file",
  imageCount: (current: number, max: number) => `${current} of ${max} images`,
  imageListAriaLabel: "Product images",
  imagePosition: (position: number, total: number) => `Gallery image ${position} of ${total}`,
  primaryImage: "Primary",
  setPrimaryImage: "Set as primary",
  setPrimaryImageFor: (name: string) => `Set ${name} as primary image`,
  noFile: "No file chosen",
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
