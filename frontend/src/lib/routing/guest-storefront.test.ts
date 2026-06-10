import { describe, expect, it } from "vitest";

import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";

import {
  allowsAuthenticatedCustomerPublicBrowsing,
  isGuestStorefrontPath,
  isPublicAuthEntryPath,
  shouldRedirectAuthenticatedPublicUser,
} from "./guest-storefront";

describe("isGuestStorefrontPath", () => {
  it("matches product listing and detail", () => {
    expect(isGuestStorefrontPath("/products")).toBe(true);
    expect(
      isGuestStorefrontPath("/products/550e8400-e29b-41d4-a716-446655440000"),
    ).toBe(true);
    expect(isGuestStorefrontPath("/cart")).toBe(false);
  });
});

describe("isPublicAuthEntryPath", () => {
  it("matches login, register, and forgot password", () => {
    expect(isPublicAuthEntryPath(ROUTE.login)).toBe(true);
    expect(isPublicAuthEntryPath(ROUTE.register)).toBe(true);
    expect(isPublicAuthEntryPath(ROUTE.forgotPassword)).toBe(true);
    expect(isPublicAuthEntryPath(ROUTE.home)).toBe(false);
  });
});

describe("allowsAuthenticatedCustomerPublicBrowsing", () => {
  it("allows home and catalog for signed-in customers", () => {
    expect(allowsAuthenticatedCustomerPublicBrowsing(ROUTE.home)).toBe(true);
    expect(allowsAuthenticatedCustomerPublicBrowsing(ROUTE.products)).toBe(
      true,
    );
    expect(allowsAuthenticatedCustomerPublicBrowsing(ROUTE.cart)).toBe(false);
  });
});

describe("shouldRedirectAuthenticatedPublicUser", () => {
  it("redirects all roles off login/register", () => {
    expect(
      shouldRedirectAuthenticatedPublicUser(ROUTE.login, USER_ROLE.customer),
    ).toBe(true);
    expect(
      shouldRedirectAuthenticatedPublicUser(ROUTE.register, USER_ROLE.manager),
    ).toBe(true);
  });

  it("keeps customers on catalog but redirects operational roles", () => {
    expect(
      shouldRedirectAuthenticatedPublicUser(ROUTE.products, USER_ROLE.customer),
    ).toBe(false);
    expect(
      shouldRedirectAuthenticatedPublicUser(ROUTE.products, USER_ROLE.admin),
    ).toBe(true);
  });
});
