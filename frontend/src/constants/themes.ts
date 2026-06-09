/** Route-scoped UI themes (set via `data-theme` on layout root). */
export const APP_THEME = {
  storefront: "storefront",
  dashboard: "dashboard",
} as const;

export type AppTheme = (typeof APP_THEME)[keyof typeof APP_THEME];
