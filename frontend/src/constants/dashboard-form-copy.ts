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

/** Catalog image upload — control copy; formats/size hint from `cloudinaryProductImageFieldHint()`. */
export const CATALOG_IMAGE_FIELD_COPY = {
  actionsAriaLabel: "Product image actions",
  chooseFile: "Choose file",
  noFile: "No file chosen",
  uploadAriaLabel: "Choose a product image file",
  uploading: "Uploading…",
  replace: "Replace",
  remove: "Remove",
} as const;
