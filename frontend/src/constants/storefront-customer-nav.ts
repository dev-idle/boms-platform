import {
  CartIcon,
  OrdersIcon,
  UserIcon,
} from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";

export const STOREFRONT_CUSTOMER_NAV = [
  { href: ROUTE.cart, label: "Cart", Icon: CartIcon },
  { href: ROUTE.orders, label: "Orders", Icon: OrdersIcon },
  { href: ROUTE.customer.account.profile, label: "Account", Icon: UserIcon },
] as const;

function isCustomerAccountPath(pathname: string): boolean {
  return (
    pathname === ROUTE.customer.account.profile ||
    pathname === ROUTE.customer.account.password ||
    pathname === ROUTE.customer.account.delete
  );
}

/** Highlights list nav and order detail routes under `/orders`. */
export function isStorefrontCustomerNavActive(
  pathname: string,
  href: string,
): boolean {
  if (pathname === href) {
    return true;
  }
  if (href === ROUTE.orders && pathname.startsWith("/orders/")) {
    return true;
  }
  if (href === ROUTE.customer.account.profile && isCustomerAccountPath(pathname)) {
    return true;
  }
  return false;
}
