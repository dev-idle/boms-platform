import { describe, expect, it } from "vitest";

import { cloudinaryDeliveryFileLabel, formatCloudinaryMaxImageSize, isDuplicateCatalogFileLabel } from "./format";
import {
  cloudinaryImageTooLargeMessage,
  cloudinaryProductImageUploadNote,
} from "./messages";

describe("formatCloudinaryMaxImageSize", () => {
  it("formats whole MiB values", () => {
    expect(formatCloudinaryMaxImageSize(5 * 1024 * 1024)).toBe("5 MB");
  });
});

describe("cloudinaryImageTooLargeMessage", () => {
  it("uses max_bytes from the API", () => {
    expect(cloudinaryImageTooLargeMessage(5 * 1024 * 1024)).toBe(
      "Image must be 5 MB or smaller",
    );
  });
});

describe("cloudinaryProductImageUploadNote", () => {
  it("uses the configured max image size", () => {
    expect(cloudinaryProductImageUploadNote(5 * 1024 * 1024)).toBe(
      "JPG, PNG, WebP, or AVIF. Up to 5 MB each, 5 images max.",
    );
  });
});

describe("cloudinaryDeliveryFileLabel", () => {
  it("uses the last URL path segment", () => {
    expect(
      cloudinaryDeliveryFileLabel(
        "https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg",
      ),
    ).toBe("loaf.jpg");
  });
});

describe("isDuplicateCatalogFileLabel", () => {
  it("detects duplicate file names case-insensitively", () => {
    expect(
      isDuplicateCatalogFileLabel(
        "Loaf.JPG",
        ["https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg"],
        {},
      ),
    ).toBe(true);
  });

  it("allows distinct file names", () => {
    expect(
      isDuplicateCatalogFileLabel(
        "croissant.jpg",
        ["https://res.cloudinary.com/demo/image/upload/v1/boms/products/loaf.jpg"],
        {},
      ),
    ).toBe(false);
  });
});
