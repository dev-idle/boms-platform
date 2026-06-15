import { afterEach, describe, expect, it, vi } from "vitest";

import { CLOUDINARY_UPLOAD_COPY } from "./messages";
import {
  type CloudinaryUploadSignature,
  validateCloudinarySignatureEnv,
} from "./signature";

const baseSignature: CloudinaryUploadSignature = {
  allowed_formats: "jpg,png,webp,avif",
  api_key: "key",
  cloud_name: "demo",
  folder: "boms/products",
  max_bytes: 5 * 1024 * 1024,
  signature: "abc",
  timestamp: 1_700_000_000,
  unique_filename: "true",
  upload_url: "https://api.cloudinary.com/v1_1/demo/image/upload",
};

describe("validateCloudinarySignatureEnv", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("passes when cloud name and folder match public env", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER", "boms/products");

    expect(() => validateCloudinarySignatureEnv(baseSignature)).not.toThrow();
  });

  it("rejects cloud name drift", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "other");

    expect(() => validateCloudinarySignatureEnv(baseSignature)).toThrow(
      CLOUDINARY_UPLOAD_COPY.cloudNameMismatch,
    );
  });

  it("rejects folder drift", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER", "boms/staging");

    expect(() => validateCloudinarySignatureEnv(baseSignature)).toThrow(
      CLOUDINARY_UPLOAD_COPY.folderMismatch,
    );
  });

  it("rejects upload URLs outside the Cloudinary API", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER", "boms/products");

    expect(() =>
      validateCloudinarySignatureEnv({
        ...baseSignature,
        upload_url: "https://evil.example/upload",
      }),
    ).toThrow(CLOUDINARY_UPLOAD_COPY.uploadFailedRemote);
  });
});
