/** Rosé/cream/berry patisserie — home hero and auth brand panel. */
const ROSE_PATISSERIE_PHOTO = "photo-1488477181946-6428a0291777";

/** Chocolate celebration cake — featured product grid fallback. */
const CHOCOLATE_CAKE_PHOTO = "photo-1578985545062-69928b1d9587";

function unsplashImageUrl(photoId: string, width: number): string {
  return `https://images.unsplash.com/${photoId}?auto=format&fit=crop&w=${width}&q=80`;
}

export const STOREFRONT_HERO_IMAGE_URL = unsplashImageUrl(ROSE_PATISSERIE_PHOTO, 1200);
export const AUTH_PATISSERIE_IMAGE_URL = unsplashImageUrl(ROSE_PATISSERIE_PHOTO, 960);
export const STOREFRONT_CHOCOLATE_PRODUCT_IMAGE_URL = unsplashImageUrl(
  CHOCOLATE_CAKE_PHOTO,
  960,
);
