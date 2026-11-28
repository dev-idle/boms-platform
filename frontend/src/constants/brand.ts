/** Customer-facing bakery brand (BOMS is the internal project name only). */
export const BRAND = {
  name: "Choux",
  tagline:
    "Handcrafted pastries and celebration cakes, made fresh daily for pickup.",
  /** Auth panel statement — the one message on the sign-in split. */
  authStatement: { lead: ["Order online,", "pick up at"], accent: "the counter." },
  /** The micro-label over the sign-in panel's statement. */
  panelLabel: "Patisserie · Baked fresh daily",
  contactEmail: "nghiantgcc19224@fpt.edu.vn",
  contactPhone: "0123 456 789",
  addressLine: "Ninh Kieu, Can Tho",
} as const;

/** `tel:` target derived once — topbar and footer must dial the same number. */
export const BRAND_PHONE_TEL_HREF = `tel:${BRAND.contactPhone.replace(/\D/g, "")}`;
