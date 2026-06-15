"use client";

import { useRef, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { DashboardImageUploadIcon } from "@/components/icons/dashboard-ui-icons";
import { CATALOG_IMAGE_FIELD_COPY } from "@/constants/dashboard-form-copy";
import { isApiError } from "@/lib/errors";
import {
  CLOUDINARY_ALLOWED_IMAGE_TYPES,
  CLOUDINARY_MAX_IMAGE_BYTES,
  isCloudinaryConfigured,
} from "@/lib/cloudinary/config";
import { fetchCloudinaryUploadSignature } from "@/lib/cloudinary/signature";
import { uploadImageToCloudinary } from "@/lib/cloudinary/upload";
import { cn } from "@/lib/utils";

type CatalogImageFieldProps = {
  disabled?: boolean;
  onChange: (next: string | null) => void;
  value: string | null;
};

export function CatalogImageField({
  disabled = false,
  onChange,
  value,
}: CatalogImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);

  if (!isCloudinaryConfigured()) {
    return null;
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }
    if (!CLOUDINARY_ALLOWED_IMAGE_TYPES.includes(file.type as (typeof CLOUDINARY_ALLOWED_IMAGE_TYPES)[number])) {
      toast.error("Use JPG, PNG, WebP, or AVIF");
      return;
    }
    if (file.size > CLOUDINARY_MAX_IMAGE_BYTES) {
      toast.error("Image must be 5 MB or smaller");
      return;
    }

    setIsUploading(true);
    try {
      const signature = await fetchCloudinaryUploadSignature();
      const secureUrl = await uploadImageToCloudinary(file, signature);
      onChange(secureUrl);
      toast.success("Image uploaded");
    } catch (error) {
      if (isApiError(error)) {
        toast.error(error.message);
      } else if (error instanceof Error && error.message) {
        toast.error(error.message);
      } else {
        toast.error("Failed to upload image");
      }
    } finally {
      setIsUploading(false);
    }
  }

  return (
    <div className="catalog-image-field">
      <input
        accept={CLOUDINARY_ALLOWED_IMAGE_TYPES.join(",")}
        className="sr-only"
        disabled={disabled || isUploading}
        onChange={handleFileChange}
        ref={inputRef}
        type="file"
      />
      {value ? (
        <div className="catalog-image-field-preview">
          {/* eslint-disable-next-line @next/next/no-img-element -- manager preview of external delivery URL */}
          <img
            alt="Product preview"
            className="catalog-image-field-preview-img"
            src={value}
          />
          <div className="catalog-image-field-actions">
            <Button
              disabled={disabled || isUploading}
              onClick={() => inputRef.current?.click()}
              type="button"
              variant="outline"
            >
              {isUploading
                ? CATALOG_IMAGE_FIELD_COPY.uploading
                : CATALOG_IMAGE_FIELD_COPY.replace}
            </Button>
            <Button
              disabled={disabled || isUploading}
              onClick={() => onChange(null)}
              type="button"
              variant="ghost"
            >
              {CATALOG_IMAGE_FIELD_COPY.remove}
            </Button>
          </div>
        </div>
      ) : (
        <button
          aria-busy={isUploading}
          aria-label={CATALOG_IMAGE_FIELD_COPY.uploadAriaLabel}
          className={cn(
            "catalog-image-field-dropzone",
            isUploading && "catalog-image-field-dropzone--uploading",
            (disabled || isUploading) && "catalog-image-field-dropzone--disabled",
          )}
          disabled={disabled || isUploading}
          onClick={() => inputRef.current?.click()}
          type="button"
        >
          <span className="catalog-image-field-dropzone-inner">
            <DashboardImageUploadIcon className="catalog-image-field-dropzone-icon" />
            <span className="catalog-image-field-dropzone-label">
              {isUploading
                ? CATALOG_IMAGE_FIELD_COPY.uploading
                : CATALOG_IMAGE_FIELD_COPY.uploadPrompt}
            </span>
          </span>
        </button>
      )}
    </div>
  );
}
