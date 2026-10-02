import type { Metadata } from "next";

/** Browser tab title segments — root layout appends brand suffix via metadata template. */
export const PAGE_TITLES = {
  home: "Home",
  shop: "Shop",
  product: "Product",
  terms: "Terms of Sale",
  privacy: "Privacy Policy",
  refundPolicy: "Refund Policy",
  signIn: "Sign in",
  createAccount: "Create account",
  forgotPassword: "Forgot password",
  resetPassword: "Choose a new password",
  verifyEmail: "Confirm your email",
  cart: "Cart",
  orders: "Orders",
  saved: "Saved",
  production: "Production",
  orderDetail: "Order Detail",
  ticket: "Ticket",
  prepQueue: "Prep Queue",
  pickups: "Pickups",
  newOrder: "New Order",
  availability: "Availability",
  messages: "Messages",
  account: "Account",
  profile: "Profile",
  changePassword: "Change Password",
  deleteAccount: "Delete Account",
  dashboard: "Dashboard",
  users: "User Management",
  settings: "Store Settings",
  newUser: "New User",
  userDetail: "User Detail",
  products: "Products",
  newProduct: "New Product",
  editProduct: "Edit Product",
  categories: "Categories",
  newCategory: "New Category",
  editCategory: "Edit Category",
  combos: "Combos",
  newCombo: "New Combo",
  editCombo: "Edit Combo",
  discountCodes: "Discount Codes",
  newDiscountCode: "New Discount Code",
  editDiscountCode: "Edit Discount Code",
  reviews: "Reviews",
  breadcrumbDetail: "Detail",
  breadcrumbNew: "New",
} as const;

/** Page title segment — root `metadata.title.template` appends ` | Choux`. */
export function pageTitle(
  title: (typeof PAGE_TITLES)[keyof typeof PAGE_TITLES] | string,
  description?: string,
): Metadata {
  return description ? { title, description } : { title };
}
