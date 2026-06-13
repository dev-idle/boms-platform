import type { Metadata } from "next";

/** Browser tab title segments — root layout appends brand suffix via metadata template. */
export const PAGE_TITLES = {
  home: "Home",
  shop: "Shop",
  product: "Product",
  signIn: "Sign In",
  createAccount: "Create Account",
  forgotPassword: "Forgot Password",
  cart: "Cart",
  orders: "Orders",
  orderDetail: "Order Detail",
  profile: "Profile",
  myProfile: "My Profile",
  changePassword: "Change Password",
  deleteAccount: "Delete Account",
  dashboard: "Dashboard",
  users: "User Management",
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
  breadcrumbDetail: "Detail",
  breadcrumbNew: "New",
} as const;

export type PageTitleKey = keyof typeof PAGE_TITLES;

/** Page title segment — root `metadata.title.template` appends ` | Choux`. */
export function pageTitle(
  title: (typeof PAGE_TITLES)[PageTitleKey] | string,
  description?: string,
): Metadata {
  return description ? { title, description } : { title };
}
