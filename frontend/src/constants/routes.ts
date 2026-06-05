/**
 * Canonical paths — single source for `src/proxy.ts`, layouts, and links.
 *
 * URL conventions (one role = one namespace):
 *   - Public:   /, /login, /register
 *   - Customer: /products, /products/:id, /cart, /orders, /customer/account/*
 *   - Staff:    /staff/account/*
 *   - Baker:    /baker/account/*
 *   - Manager:  /manager/categories, /manager/products, /manager/combos,
 *               /manager/discount-codes, /manager/account/*
 *   - Admin:    /admin, /admin/users, /admin/account/profile (password on profile page)
 */
export const ROUTE = {
  home: "/",
  login: "/login",
  register: "/register",
  products: "/products",
  productDetail: (id: string) => `/products/${id}`,
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
  baker: {
    account: {
      profile: "/baker/account/profile",
      password: "/baker/account/password",
    },
  },
  manager: {
    categories: "/manager/categories",
    categoriesNew: "/manager/categories/new",
    categoryDetail: (id: string) => `/manager/categories/${id}`,
    products: "/manager/products",
    productsNew: "/manager/products/new",
    productDetail: (id: string) => `/manager/products/${id}`,
    combos: "/manager/combos",
    combosNew: "/manager/combos/new",
    comboDetail: (id: string) => `/manager/combos/${id}`,
    discountCodes: "/manager/discount-codes",
    discountCodesNew: "/manager/discount-codes/new",
    discountCodeDetail: (id: string) => `/manager/discount-codes/${id}`,
    account: {
      profile: "/manager/account/profile",
      password: "/manager/account/password",
    },
  },
  admin: {
    dashboard: "/admin",
    users: "/admin/users",
    usersNew: "/admin/users/new",
    userDetail: (id: string) => `/admin/users/${id}`,
    account: {
      /** Self-service profile + password change (no separate password URL). */
      profile: "/admin/account/profile",
    },
  },
} as const;

export const CUSTOMER_ROUTE_PREFIXES = [
  ROUTE.products,
  ROUTE.cart,
  ROUTE.orders,
  "/customer",
] as const;

export const STAFF_ROUTE_PREFIXES = ["/staff"] as const;

export const BAKER_ROUTE_PREFIXES = ["/baker"] as const;

export const MANAGER_ROUTE_PREFIXES = ["/manager"] as const;

export const ADMIN_ROUTE_PREFIXES = ["/admin"] as const;

export const PROTECTED_ROUTE_PREFIXES = [
  ...CUSTOMER_ROUTE_PREFIXES,
  ...STAFF_ROUTE_PREFIXES,
  ...BAKER_ROUTE_PREFIXES,
  ...MANAGER_ROUTE_PREFIXES,
  ...ADMIN_ROUTE_PREFIXES,
] as const;

export type ProtectedRoutePrefix = (typeof PROTECTED_ROUTE_PREFIXES)[number];
