import { describe, expect, it, vi } from "vitest";

import {
  productImageUrlsResponseSchema,
  productImageUrlsSchema,
} from "./cloudinary";

describe("productImageUrlsResponseSchema", () => {
  it("accepts legacy non-Cloudinary URLs on read", () => {
    const parsed = productImageUrlsResponseSchema.parse([
      "https://cdn.example.com/loaf.jpg",
    ]);
    expect(parsed).toEqual(["https://cdn.example.com/loaf.jpg"]);
  });

  it("defaults missing or null to an empty array", () => {
    expect(productImageUrlsResponseSchema.parse(undefined)).toEqual([]);
    expect(productImageUrlsResponseSchema.parse(null)).toEqual([]);
    expect(productImageUrlsResponseSchema.parse([])).toEqual([]);
  });
});

describe("productImageUrlsSchema", () => {
  it("rejects URLs outside the configured Cloudinary folder on write", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");
    vi.stubEnv(
      "NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER",
      "boms/products",
    );

    const result = productImageUrlsSchema.safeParse([
      "https://cdn.example.com/loaf.jpg",
    ]);

    expect(result.success).toBe(false);
    vi.unstubAllEnvs();
  });
});
