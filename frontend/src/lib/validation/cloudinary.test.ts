import { describe, expect, it, vi } from "vitest";
import { z } from "zod";

import {
  catalogImageUrlResponseSchema,
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

  // The API omits `image_urls` for a product without images. Zod 4 treats an
  // object key as required unless its schema is optional, so this must hold
  // inside an object, not only for a bare `undefined`.
  it("accepts a response that omits the key", () => {
    const product = z.object({ id: z.string(), image_urls: productImageUrlsResponseSchema });
    expect(product.parse({ id: "p1" })).toEqual({ id: "p1", image_urls: [] });
  });
});

describe("catalogImageUrlResponseSchema", () => {
  it("reads a missing, null or present image", () => {
    const combo = z.object({ id: z.string(), image_url: catalogImageUrlResponseSchema });
    expect(combo.parse({ id: "c1" })).toEqual({ id: "c1", image_url: null });
    expect(combo.parse({ id: "c1", image_url: null })).toEqual({ id: "c1", image_url: null });
    expect(
      combo.parse({ id: "c1", image_url: "https://cdn.example.com/box.jpg" }),
    ).toEqual({ id: "c1", image_url: "https://cdn.example.com/box.jpg" });
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
