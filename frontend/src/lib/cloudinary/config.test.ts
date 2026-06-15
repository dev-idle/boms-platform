import { describe, expect, it } from "vitest";

import {
  catalogProductImageUrl,
  isCloudinaryDeliveryUrl,
  isCloudinaryDeliveryUrlInFolder,
} from "./config";

describe("isCloudinaryDeliveryUrl", () => {
  it("accepts delivery URLs for the configured cloud", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg";
    expect(isCloudinaryDeliveryUrl(url, "demo")).toBe(true);
  });

  it("rejects other hosts and clouds", () => {
    expect(
      isCloudinaryDeliveryUrl(
        "https://res.cloudinary.com/other/image/upload/x.jpg",
        "demo",
      ),
    ).toBe(false);
    expect(isCloudinaryDeliveryUrl("https://example.com/x.jpg", "demo")).toBe(
      false,
    );
  });
});

describe("isCloudinaryDeliveryUrlInFolder", () => {
  it("accepts URLs under the product upload folder", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_640/v1/boms/products/loaf.jpg";
    expect(isCloudinaryDeliveryUrlInFolder(url, "boms/products", "demo")).toBe(
      true,
    );
  });

  it("rejects URLs outside the product upload folder", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/v1/boms/other/loaf.jpg";
    expect(isCloudinaryDeliveryUrlInFolder(url, "boms/products", "demo")).toBe(
      false,
    );
  });

  it("rejects path traversal outside the product upload folder", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/v1/boms/products/../other/loaf.jpg";
    expect(isCloudinaryDeliveryUrlInFolder(url, "boms/products", "demo")).toBe(
      false,
    );
  });
});

describe("catalogProductImageUrl", () => {
  it("adds transforms for Cloudinary URLs", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg";
    expect(catalogProductImageUrl(url, 640)).toBe(
      "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_640/boms/products/loaf.jpg",
    );
  });

  it("replaces existing transforms instead of nesting them", () => {
    const url =
      "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_640/v1/boms/products/loaf.jpg";
    expect(catalogProductImageUrl(url, 960)).toBe(
      "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_960/boms/products/loaf.jpg",
    );
  });

  it("returns external URLs unchanged", () => {
    const url = "https://images.example.test/loaf.jpg";
    expect(catalogProductImageUrl(url, 640)).toBe(url);
  });
});
