/** Customer-facing bakery brand (BOMS is the internal project name only). */
export const BRAND = {
  name: "Choux",
  tagline:
    "Handcrafted pastries and celebration cakes, made fresh daily for pickup.",
  /** Auth panel statement — the one message on the sign-in split. */
  authStatement: { lead: ["Baked the", "morning you"], accent: "collect." },
  establishedLabel: "Patisserie · Est. 2012",
  contactEmail: "hello@chouxbakery.example",
  contactPhone: "(555) 123-4567",
  addressLine: "123 Greige Lane",
  /** Pickup window — one source for the hero stat, auth panel, and footer. */
  pickupHours: "8:00 — 23:00",
  pickupHoursCompact: "8–23",
  pickupHoursFooter: "Every day · 8:00 — 23:00",
  kitchenCloseNote: "Kitchen closes 22:00",
  /** Display stat on the home hero — years since establishment. */
  yearsInBusiness: 14,
} as const;

/** `tel:` target derived once — topbar and footer must dial the same number. */
export const BRAND_PHONE_TEL_HREF = `tel:${BRAND.contactPhone.replace(/\D/g, "")}`;
