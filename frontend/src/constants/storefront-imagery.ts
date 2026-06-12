/**
 * Storefront imagery — art direction SSOT (Choux Matcha v3 §2)
 *
 * Every hero + product-card image MUST satisfy ALL of:
 * 1. SUBJECT — pastry in brand range (matcha / pistachio / cream / vanilla / light brown).
 *    Strong red/pink only as a detail (<10% frame); berry shots are product-grid fallbacks only.
 * 2. KEY — high-key: bright backdrop (white / cream / light marble / linen). No black backgrounds,
 *    no dark moody grading, no low-key studio lighting.
 * 3. LUMINANCE — average brightness near porcelain bg (#FBFAF9). No “dark holes” on the page.
 *    Quick test: scale thumbnail to ~50px; if noticeably darker than page bg → reject.
 *
 * Suggested Unsplash search keywords:
 * - "matcha cream puff white background"
 * - "pistachio choux pastry bright"
 * - "matcha eclair cream backdrop flat lay"
 */

/** Self-hosted hero — matcha cream puffs, high-key (curated reference shot). */
const HERO_MATCHA_CREAM_PUFFS_PATH = "/storefront/hero-matcha-cream-puffs.webp";

/** Vanilla cream puffs on porcelain — light wood, high-key. */
const VANILLA_CREAM_CHOUX_PHOTO = "photo-1707578365452-56abc4c38873";

/** Pistachio profiteroles on a white plate — bright outdoor, high-key. */
const PISTACHIO_PROFITEROLES_PHOTO = "photo-1749280446565-7c6728609f0e";

/** White-glazed eclair — cream/vanilla choux, bright food styling. */
const VANILLA_ECLAIR_PHOTO = "photo-1774119711073-36b8cc2490a2";

/** Berry patisserie display — product grid only; red accent kept small in frame. */
const BERRY_PATISSERIE_PHOTO = "photo-1488477181946-6428a0291777";

type UnsplashCrop = {
  width: number;
  height?: number;
};

function unsplashImageUrl(photoId: string, { width, height }: UnsplashCrop): string {
  const params = new URLSearchParams({
    auto: "format",
    fit: "crop",
    w: String(width),
    q: "80",
  });
  if (height !== undefined) {
    params.set("h", String(height));
  }
  return `https://images.unsplash.com/${photoId}?${params}`;
}

export const STOREFRONT_HERO_IMAGE_URL = HERO_MATCHA_CREAM_PUFFS_PATH;
export const AUTH_PATISSERIE_IMAGE_URL = HERO_MATCHA_CREAM_PUFFS_PATH;
export const STOREFRONT_MATCHA_PRODUCT_IMAGE_URL = HERO_MATCHA_CREAM_PUFFS_PATH;

export const STOREFRONT_BRAND_STORY_IMAGE_URL = unsplashImageUrl(VANILLA_ECLAIR_PHOTO, {
  width: 900,
  height: 900,
});

export const STOREFRONT_VANILLA_CHOUX_PRODUCT_IMAGE_URL = unsplashImageUrl(
  VANILLA_CREAM_CHOUX_PHOTO,
  { width: 960 },
);
export const STOREFRONT_PISTACHIO_PRODUCT_IMAGE_URL = unsplashImageUrl(
  PISTACHIO_PROFITEROLES_PHOTO,
  { width: 960 },
);
export const STOREFRONT_BERRY_PRODUCT_IMAGE_URL = unsplashImageUrl(BERRY_PATISSERIE_PHOTO, {
  width: 960,
});

const CATALOG_FALLBACK_IMAGES = [
  STOREFRONT_MATCHA_PRODUCT_IMAGE_URL,
  STOREFRONT_PISTACHIO_PRODUCT_IMAGE_URL,
  STOREFRONT_VANILLA_CHOUX_PRODUCT_IMAGE_URL,
  STOREFRONT_BERRY_PRODUCT_IMAGE_URL,
] as const;

/** Rotating high-key fallbacks when a catalog item has no manager-provided image. */
export function catalogProductFallbackImageUrl(index: number): string {
  return CATALOG_FALLBACK_IMAGES[index % CATALOG_FALLBACK_IMAGES.length];
}
