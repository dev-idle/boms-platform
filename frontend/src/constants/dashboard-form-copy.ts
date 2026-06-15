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
  operationalEmployeeCode: "Permanent staff ID. Assigned at creation.",
  productImageCloudinary: "JPG, PNG, WebP, or AVIF up to 5 MB.",
  productImageUrlFallback: "Paste an HTTPS image URL when Cloudinary is not configured.",
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

/** Catalog image upload affordance — inside the dropzone, not FieldControl hints. */
export const CATALOG_IMAGE_FIELD_COPY = {
  uploadAriaLabel: "Choose a product image to upload",
  uploadPrompt: "Choose a product image",
  uploading: "Uploading image…",
  replace: "Replace image",
  remove: "Remove",
} as const;
