import { describe, expect, it } from "vitest";

import { ROUTE } from "@/constants/routes";
import { USER_ROLE } from "@/constants/roles";

import { parseRoleHint, signedInLanding } from "./signed-in-landing";

describe("parseRoleHint", () => {
  it("accepts the five roles", () => {
    for (const role of Object.values(USER_ROLE)) {
      expect(parseRoleHint(role), role).toBe(role);
    }
  });

  it("ignores anything else, so a tampered cookie routes nobody", () => {
    for (const value of [undefined, "", "root", "ADMIN", "admin ", "manager;admin"]) {
      expect(parseRoleHint(value), String(value)).toBeUndefined();
    }
  });
});

describe("signedInLanding", () => {
  it("sends a signed-in visitor away from the sign-in pages", () => {
    expect(signedInLanding(ROUTE.login, "", USER_ROLE.manager)).toBe(ROUTE.manager.dashboard);
    expect(signedInLanding(ROUTE.register, "", USER_ROLE.baker)).toBe(ROUTE.baker.production);
  });

  it("honours a deep link the visitor followed before signing in", () => {
    expect(
      signedInLanding(ROUTE.login, "?next=%2Fmanager%2Fproducts", USER_ROLE.manager),
    ).toBe("/manager/products");
  });

  it("ignores a deep link into someone else's area", () => {
    expect(signedInLanding(ROUTE.login, "?next=%2Fadmin%2Fusers", USER_ROLE.manager)).toBe(
      ROUTE.manager.dashboard,
    );
    expect(signedInLanding(ROUTE.login, "?next=https%3A%2F%2Fevil.test", USER_ROLE.manager)).toBe(
      ROUTE.manager.dashboard,
    );
  });

  it("keeps staff out of the storefront, where they have nothing to do", () => {
    // The wait for the session is what used to make this a visible jump.
    expect(signedInLanding(ROUTE.home, "", USER_ROLE.manager)).toBe(ROUTE.manager.dashboard);
    expect(signedInLanding(ROUTE.products, "", USER_ROLE.staff)).toBe(
      ROUTE.staff.account.profile,
    );
    expect(signedInLanding("/products/almond-croissant", "", USER_ROLE.admin)).toBe(
      ROUTE.admin.dashboard,
    );
  });

  it("lets a customer browse the shop while signed in", () => {
    for (const path of [ROUTE.home, ROUTE.products, "/products/almond-croissant"]) {
      expect(signedInLanding(path, "", USER_ROLE.customer), path).toBeNull();
    }
  });

  it("sends a visitor out of an area that is not theirs", () => {
    expect(signedInLanding("/admin/users", "", USER_ROLE.manager)).toBe(ROUTE.manager.dashboard);
    expect(signedInLanding(ROUTE.cart, "", USER_ROLE.manager)).toBe(ROUTE.manager.dashboard);
  });

  it("leaves a visitor alone inside their own area", () => {
    expect(signedInLanding("/manager/products/new", "", USER_ROLE.manager)).toBeNull();
    expect(signedInLanding(ROUTE.cart, "", USER_ROLE.customer)).toBeNull();
  });
});

describe("the edge only lands document navigations", () => {
  // Both blocking cases the review found: a hint that disagrees with the real
  // session would otherwise bounce against the gate that corrects it, and the
  // sign-in page would be unreachable exactly when the session cannot restore.
  it("still answers for the paths the gates would fight over", () => {
    expect(signedInLanding(ROUTE.login, "", USER_ROLE.manager)).toBe(
      ROUTE.manager.dashboard,
    );
    expect(signedInLanding("/admin/users", "", USER_ROLE.manager)).toBe(
      ROUTE.manager.dashboard,
    );
  });

  it("never lands an API path, whatever the hint claims", () => {
    for (const path of ["/api/v1/me", "/api/v1/auth/refresh", "/api/v1/manager/products"]) {
      expect(signedInLanding(path, "", USER_ROLE.manager), path).toBeNull();
    }
  });
});
