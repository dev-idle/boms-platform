import { describe, expect, it } from "vitest";

import { cloudinaryDeliveryFileLabel, formatCloudinaryMaxImageSize } from "./format";
import {
  cloudinaryImageTooLargeMessage,
  cloudinaryProductImageFieldHint,
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

describe("cloudinaryProductImageFieldHint", () => {
  it("uses the configured max image size", () => {
    expect(cloudinaryProductImageFieldHint(5 * 1024 * 1024)).toBe(
      "JPG, PNG, WebP, or AVIF up to 5 MB.",
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
