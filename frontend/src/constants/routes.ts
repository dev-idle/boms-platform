/**
 * Canonical paths — single source for `src/proxy.ts`, layouts, and links.
 */
export const ROUTE = {
  home: "/",
  login: "/login",
  register: "/register",
  products: "/products",
  cart: "/cart",
  orders: "/orders",
  admin: {
    dashboard: "/dashboard",
    products: "/dashboard/products",
    orders: "/dashboard/orders",
    users: "/dashboard/users",
  },
} as const;

export const PROTECTED_ROUTE_PREFIXES = [
  ROUTE.products,
  ROUTE.cart,
  ROUTE.orders,
  ROUTE.admin.dashboard,
] as const;

export const PUBLIC_AUTH_ROUTE_PREFIXES = [
  ROUTE.login,
  ROUTE.register,
] as const;

export type ProtectedRoutePrefix = (typeof PROTECTED_ROUTE_PREFIXES)[number];
export type PublicAuthRoutePrefix = (typeof PUBLIC_AUTH_ROUTE_PREFIXES)[number];
