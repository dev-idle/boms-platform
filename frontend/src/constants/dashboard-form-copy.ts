/**
 * Dashboard form copy — hints and switch labels (SSOT).
 *
 * Hints (`FORM_FIELD_HINT`): text fields only — between label and control.
 *
 * Switch hints (`FORM_SWITCH_HINT`): under the toggle label, left column.
 * State the off-state effect; do not paraphrase the label.
 */

export const FORM_FIELD_HINT = {
  catalogPrice: "Shown to customers exactly as entered.",
  catalogSlugCreate: "Auto-filled from name.",
  catalogSortOrder: "Lower values appear first. Leave empty for 0.",
  comboItems: "Each product once. Add two products, or one product with quantity at least 2.",
  discountMaxUses: "Leave empty for unlimited uses.",
  discountMinOrder: "Minimum cart total. Leave empty for no minimum.",
  discountMaxDiscount: "Caps a percent discount. Leave empty for no cap.",
  discountPercentOff: "Whole number from 1 to 100.",
  comboImageUrlFallback: "One HTTPS image URL.",
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
  /** A field that takes exactly one image never offers to add several. */
  addOneImage: "Add image",
  /** Reasons a row action is greyed rather than removed. */
  alreadyPrimary: "This is already the primary image.",
  busy: "Wait for the upload to finish.",
  previewUnavailable: "Enter an image URL to preview it.",
  removeImage: (name: string) => `Remove ${name}`,
  imageCount: (current: number, max: number) => `${current} of ${max} images`,
  imageListAriaLabel: "Product images",
  /** The badge marking the cover image is decorative, so the row says it. */
  imagePosition: (position: number, total: number, primary: boolean) =>
    primary
      ? `Gallery image ${position} of ${total}, the primary image`
      : `Gallery image ${position} of ${total}`,
  /**
   * The action that makes an image the cover. One word, because the row is the
   * busiest in the dashboard and the accessible name below carries the verb.
   */
  primaryImage: "Primary",
  setPrimaryImageFor: (name: string) => `Set ${name} as the primary image`,
  uploadAriaLabel: "Choose product image files",
  uploadProgress: "Uploading images…",
  remove: "Remove",
  view: "View",
  viewImage: (name: string) => `View ${name}`,
  viewImageDialogTitle: "Image preview",
  closePreview: "Close",
  previewLoadError: "Failed to load image preview.",
  previewLoading: "Loading image…",
} as const;
