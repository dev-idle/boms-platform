/**
 * Canonical paths — single source for `src/proxy.ts`, layouts, and links.
 *
 * URL conventions:
 *   - Public:   /, /login, /register
 *   - Customer: /products, /cart, /orders, /customer/account/*
 *   - Staff:    /staff/account/*
 *   - Admin:    /admin, /admin/products, /admin/orders, /admin/users, /admin/account/*
 */
export const ROUTE = {
  home: "/",
  login: "/login",
  register: "/register",
  products: "/products",
  cart: "/cart",
  orders: "/orders",
  customer: {
    account: {
      profile: "/customer/account/profile",
      password: "/customer/account/password",
      delete: "/customer/account/delete",
    },
  },
  staff: {
    account: {
      profile: "/staff/account/profile",
      password: "/staff/account/password",
    },
  },
  admin: {
    dashboard: "/admin",
    products: "/admin/products",
    orders: "/admin/orders",
    users: "/admin/users",
    usersNew: "/admin/users/new",
    userDetail: (id: string) => `/admin/users/${id}`,
    account: {
      profile: "/admin/account/profile",
      password: "/admin/account/password",
    },
  },
} as const;

export const ADMIN_ROUTE_PREFIXES = ["/admin"] as const;

export const CUSTOMER_ROUTE_PREFIXES = [
  ROUTE.products,
  ROUTE.cart,
  ROUTE.orders,
  "/customer",
] as const;

export const STAFF_ROUTE_PREFIXES = ["/staff"] as const;

export const PROTECTED_ROUTE_PREFIXES = [
  ...CUSTOMER_ROUTE_PREFIXES,
  ...STAFF_ROUTE_PREFIXES,
  ...ADMIN_ROUTE_PREFIXES,
] as const;

export type ProtectedRoutePrefix = (typeof PROTECTED_ROUTE_PREFIXES)[number];
