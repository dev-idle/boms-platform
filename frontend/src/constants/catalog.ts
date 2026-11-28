/**
 * Storefront catalog page sizes — one value per list, read by the Zod filter
 * defaults in `lib/schemas/catalog.ts` and by the components that request a
 * page. `lib/` cannot import `features/*`, so the number lives here.
 */

/** Categories fit on one page; the shop chips and the header menu show them all. */
export const CATALOG_CATEGORIES_PAGE_SIZE = 100;

/** Products per Shop page — the API's cap, so the shop lists every product in
 *  one section per category; pages appear only past it. */
export const CATALOG_PRODUCTS_PAGE_SIZE = 100;

/** Combos per page. The shop chips read the same page to know the section
 *  exists, so both share one query instead of two round trips for one answer. */
export const CATALOG_COMBOS_PAGE_SIZE = 12;
