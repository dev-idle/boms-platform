import { describe, expect, it } from "vitest";

import { ROUTE } from "@/constants/routes";
import { isStorefrontCustomerNavActive } from "@/constants/storefront-customer-nav";

describe("isStorefrontCustomerNavActive", () => {
  it("matches exact href", () => {
    expect(isStorefrontCustomerNavActive("/cart", ROUTE.cart)).toBe(true);
    expect(isStorefrontCustomerNavActive("/cart", ROUTE.orders)).toBe(false);
  });

  it("highlights orders on order detail routes", () => {
    expect(
      isStorefrontCustomerNavActive("/orders/abc", ROUTE.orders),
    ).toBe(true);
  });

  it("marks account active on profile, password, and delete routes", () => {
    expect(
      isStorefrontCustomerNavActive(
        ROUTE.customer.account.profile,
        ROUTE.customer.account.profile,
      ),
    ).toBe(true);
    expect(
      isStorefrontCustomerNavActive(
        ROUTE.customer.account.password,
        ROUTE.customer.account.profile,
      ),
    ).toBe(true);
    expect(
      isStorefrontCustomerNavActive(
        ROUTE.customer.account.delete,
        ROUTE.customer.account.profile,
      ),
    ).toBe(true);
  });
});
