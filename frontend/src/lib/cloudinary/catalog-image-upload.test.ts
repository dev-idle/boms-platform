import { describe, expect, it } from "vitest";

import { planCatalogImageUpload } from "./catalog-image-upload";
import { CLOUDINARY_UPLOAD_COPY } from "./messages";

function jpegFile(name: string, sizeBytes = 1024): File {
  return new File([new Uint8Array(sizeBytes)], name, { type: "image/jpeg" });
}

describe("planCatalogImageUpload", () => {
  it("rejects when selection exceeds remaining gallery slots", () => {
    const result = planCatalogImageUpload({
      existingUrls: ["https://res.cloudinary.com/demo/a.jpg"],
      files: [jpegFile("b.jpg"), jpegFile("c.jpg"), jpegFile("d.jpg"), jpegFile("e.jpg"), jpegFile("f.jpg")],
      labelsByUrl: {},
      maxImages: 5,
    });

    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toContain("4 slots remain");
    }
  });

  it("rejects duplicate filenames within the same selection", () => {
    const result = planCatalogImageUpload({
      existingUrls: [],
      files: [jpegFile("hero.jpg"), jpegFile("HERO.JPG")],
      labelsByUrl: {},
    });

    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.message).toBe(CLOUDINARY_UPLOAD_COPY.duplicateFileName);
    }
  });

  it("accepts a batch that fits within the gallery limit", () => {
    const files = [jpegFile("a.jpg"), jpegFile("b.jpg")];
    const result = planCatalogImageUpload({
      existingUrls: ["https://res.cloudinary.com/demo/c.jpg"],
      files,
      labelsByUrl: {},
      maxImages: 5,
    });

    expect(result).toEqual({ ok: true, files });
  });
});
