/**
 * Canonical paths — single source for `src/proxy.ts`, layouts, and links.
 *
 * URL conventions (one role = one namespace):
 *   - Public:   /, /login, /register, /forgot-password, /reset-password, /verify-email, /unsubscribe,
 *               /products, /products/:id, /terms, /privacy, /refund-policy
 *   - Customer: /cart, /orders, /customer/saved, /customer/account/*
 *   - Staff:    /staff/orders, /staff/orders/new, /staff/orders/:id, /staff/pickups, /staff/prep, /staff/availability, /staff/chat, /staff/chat/:orderId, /staff/account/*
 *   - Baker:    /baker/production, /baker/production/:id, /baker/account/*
 *   - Manager:  /manager, /manager/categories, /manager/products, /manager/combos,
 *               /manager/discount-codes, /manager/reviews, /manager/promotions, /manager/promotions/new,
 *               /manager/incidents, /manager/engagement,
 *               /manager/account/*
 *   - Admin:    /admin, /admin/users, /admin/settings, /admin/account/*
 */
export const ROUTE = {
  home: "/",
  login: "/login",
  register: "/register",
  forgotPassword: "/forgot-password",
  /** Emailed links land here; the token rides in the URL fragment. */
  resetPassword: "/reset-password",
  verifyEmail: "/verify-email",
  /** Promotion emails link here; the token rides in the fragment. */
  unsubscribe: "/unsubscribe",
  products: "/products",
  productDetail: (id: string) => `/products/${id}`,
  terms: "/terms",
  privacy: "/privacy",
  refundPolicy: "/refund-policy",
  cart: "/cart",
  orders: "/orders",
  orderDetail: (id: string) => `/orders/${id}`,
  customer: {
    saved: "/customer/saved",
    account: {
      profile: "/customer/account/profile",
      password: "/customer/account/password",
      delete: "/customer/account/delete",
    },
  },
  staff: {
    orders: "/staff/orders",
    newOrder: "/staff/orders/new",
    orderDetail: (id: string) => `/staff/orders/${id}`,
    pickups: "/staff/pickups",
    prep: "/staff/prep",
    availability: "/staff/availability",
    chat: "/staff/chat",
    chatThread: (orderId: string) => `/staff/chat/${orderId}`,
    account: {
      root: "/staff/account",
      profile: "/staff/account/profile",
      password: "/staff/account/password",
    },
  },
  baker: {
    production: "/baker/production",
    productionDetail: (id: string) => `/baker/production/${id}`,
    account: {
      root: "/baker/account",
      profile: "/baker/account/profile",
      password: "/baker/account/password",
    },
  },
  manager: {
    dashboard: "/manager",
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
    reviews: "/manager/reviews",
    promotions: "/manager/promotions",
    promotionsNew: "/manager/promotions/new",
    incidents: "/manager/incidents",
    engagement: "/manager/engagement",
    account: {
      root: "/manager/account",
      profile: "/manager/account/profile",
      password: "/manager/account/password",
    },
  },
  admin: {
    dashboard: "/admin",
    users: "/admin/users",
    usersNew: "/admin/users/new",
    userDetail: (id: string) => `/admin/users/${id}`,
    settings: "/admin/settings",
    account: {
      root: "/admin/account",
      profile: "/admin/account/profile",
      password: "/admin/account/password",
    },
  },
} as const;

/** Guest-accessible storefront browse (no session required). */
export const GUEST_STOREFRONT_ROUTE_PREFIXES = [ROUTE.products] as const;

/** Customer session required (cart, orders, account). */
const CUSTOMER_PROTECTED_ROUTE_PREFIXES = [
  ROUTE.cart,
  ROUTE.orders,
  "/customer",
] as const;

export const CUSTOMER_ROUTE_PREFIXES = [
  ...GUEST_STOREFRONT_ROUTE_PREFIXES,
  ...CUSTOMER_PROTECTED_ROUTE_PREFIXES,
] as const;

export const STAFF_ROUTE_PREFIXES = ["/staff"] as const;

export const BAKER_ROUTE_PREFIXES = ["/baker"] as const;

export const MANAGER_ROUTE_PREFIXES = ["/manager"] as const;

export const ADMIN_ROUTE_PREFIXES = ["/admin"] as const;

export const PROTECTED_ROUTE_PREFIXES = [
  ...CUSTOMER_PROTECTED_ROUTE_PREFIXES,
  ...STAFF_ROUTE_PREFIXES,
  ...BAKER_ROUTE_PREFIXES,
  ...MANAGER_ROUTE_PREFIXES,
  ...ADMIN_ROUTE_PREFIXES,
] as const;
