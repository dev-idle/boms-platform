/** Curated Unsplash patisserie photo — shared by home hero and auth brand panel. */
const PATISSERIE_HERO_PHOTO = "photo-1578985545062-69928b1d9587";

function patisserieImageUrl(width: number): string {
  return `https://images.unsplash.com/${PATISSERIE_HERO_PHOTO}?auto=format&fit=crop&w=${width}&q=80`;
}

export const STOREFRONT_HERO_IMAGE_URL = patisserieImageUrl(1200);
export const AUTH_PATISSERIE_IMAGE_URL = patisserieImageUrl(960);
