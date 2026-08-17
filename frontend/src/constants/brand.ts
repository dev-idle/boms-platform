/** Customer-facing bakery brand (BOMS is the internal project name only). */
export const BRAND = {
  name: "Choux",
  tagline:
    "Handcrafted pastries and celebration cakes, made fresh daily for pickup.",
  contactEmail: "hello@chouxbakery.example",
  contactPhone: "(555) 123-4567",
  addressLine: "123 Greige Lane",
} as const;

/** `tel:` target derived once — topbar and footer must dial the same number. */
export const BRAND_PHONE_TEL_HREF = `tel:${BRAND.contactPhone.replace(/\D/g, "")}`;
