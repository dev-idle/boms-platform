import { afterEach, describe, expect, it, vi } from "vitest";

import { getCloudinaryUploadFolder } from "./config";
import { CLOUDINARY_UPLOAD_COPY } from "./messages";
import {
  type CloudinaryUploadSignature,
  validateCloudinarySignatureEnv,
} from "./signature";

const baseSignature: CloudinaryUploadSignature = {
  api_key: "key",
  cloud_name: "demo",
  folder: "boms/products",
  max_bytes: 5 * 1024 * 1024,
  params: { allowed_formats: "jpg,png,webp,avif", folder: "boms/products", timestamp: "1700000000", unique_filename: "true" },
  signature: "abc",
  upload_url: "https://api.cloudinary.com/v1_1/demo/image/upload",
};

describe("validateCloudinarySignatureEnv", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("passes when cloud name and folder match public env", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER", "boms/products");

    expect(() => validateCloudinarySignatureEnv(baseSignature, getCloudinaryUploadFolder())).not.toThrow();
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

    expect(() => validateCloudinarySignatureEnv(baseSignature, getCloudinaryUploadFolder())).toThrow(
      CLOUDINARY_UPLOAD_COPY.folderMismatch,
    );
  });

  it("leaves a folder the API names to the API", () => {
    vi.stubEnv("NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME", "demo");

    expect(() =>
      validateCloudinarySignatureEnv({ ...baseSignature, folder: "boms/references/7c1f" }),
    ).not.toThrow();
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
