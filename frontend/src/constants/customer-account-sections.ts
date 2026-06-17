import { ROUTE } from "@/constants/routes";

/** In-page anchors on the unified customer account profile route. */
export const CUSTOMER_ACCOUNT_SECTION = {
  profile: "customer-profile",
  password: "customer-password",
  delete: "customer-delete-account",
} as const;

export function customerAccountSectionHref(
  section: keyof typeof CUSTOMER_ACCOUNT_SECTION,
): string {
  return `${ROUTE.customer.account.profile}#${CUSTOMER_ACCOUNT_SECTION[section]}`;
}
