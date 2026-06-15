"use client";

import { useRef, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CATALOG_IMAGE_FIELD_COPY } from "@/constants/dashboard-form-copy";
import { isApiError } from "@/lib/errors";
import {
  CLOUDINARY_ALLOWED_IMAGE_TYPES,
  CLOUDINARY_MAX_IMAGE_BYTES,
  isCloudinaryConfigured,
} from "@/lib/cloudinary/config";
import { cloudinaryDeliveryFileLabel } from "@/lib/cloudinary/format";
import {
  CLOUDINARY_UPLOAD_COPY,
  cloudinaryImageTooLargeMessage,
} from "@/lib/cloudinary/messages";
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
  const [uploadedFileName, setUploadedFileName] = useState<string | null>(null);

  if (!isCloudinaryConfigured()) {
    return null;
  }

  const displayFileName = value
    ? (uploadedFileName ?? cloudinaryDeliveryFileLabel(value))
    : null;
  const isDisabled = disabled || isUploading;

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }
    if (!CLOUDINARY_ALLOWED_IMAGE_TYPES.includes(file.type as (typeof CLOUDINARY_ALLOWED_IMAGE_TYPES)[number])) {
      toast.error(CLOUDINARY_UPLOAD_COPY.allowedFormats);
      return;
    }
    if (file.size > CLOUDINARY_MAX_IMAGE_BYTES) {
      toast.error(cloudinaryImageTooLargeMessage(CLOUDINARY_MAX_IMAGE_BYTES));
      return;
    }

    setIsUploading(true);
    try {
      const signature = await fetchCloudinaryUploadSignature();
      const secureUrl = await uploadImageToCloudinary(file, signature);
      setUploadedFileName(file.name);
      onChange(secureUrl);
      toast.success(CLOUDINARY_UPLOAD_COPY.uploadSuccess);
    } catch (error) {
      const message = isApiError(error)
        ? error.message
        : error instanceof Error && error.message
          ? error.message
          : CLOUDINARY_UPLOAD_COPY.uploadFailed;
      toast.error(message);
    } finally {
      setIsUploading(false);
    }
  }

  function handleRemove(): void {
    setUploadedFileName(null);
    onChange(null);
  }

  function openFilePicker(): void {
    inputRef.current?.click();
  }

  return (
    <div
      className={cn(
        "catalog-image-field",
        isUploading && "catalog-image-field--uploading",
        isDisabled && "catalog-image-field--disabled",
      )}
    >
      <input
        accept={CLOUDINARY_ALLOWED_IMAGE_TYPES.join(",")}
        aria-label={CATALOG_IMAGE_FIELD_COPY.uploadAriaLabel}
        className="sr-only"
        disabled={isDisabled}
        onChange={handleFileChange}
        ref={inputRef}
        type="file"
      />
      <div className="catalog-image-field-row">
        {displayFileName ? (
          <>
            <div
              aria-label={CATALOG_IMAGE_FIELD_COPY.actionsAriaLabel}
              className="catalog-image-field-actions"
              role="toolbar"
            >
              <Button
                disabled={isDisabled}
                onClick={openFilePicker}
                size="sm"
                type="button"
                variant="outline"
              >
                {isUploading
                  ? CATALOG_IMAGE_FIELD_COPY.uploading
                  : CATALOG_IMAGE_FIELD_COPY.replace}
              </Button>
              <Button
                disabled={isDisabled}
                onClick={handleRemove}
                size="sm"
                type="button"
                variant="ghost"
              >
                {CATALOG_IMAGE_FIELD_COPY.remove}
              </Button>
            </div>
            <span
              className="catalog-image-field-name"
              role="status"
              title={displayFileName}
            >
              {displayFileName}
            </span>
          </>
        ) : (
          <>
            <Button
              disabled={isDisabled}
              onClick={openFilePicker}
              size="sm"
              type="button"
              variant="outline"
            >
              {isUploading
                ? CATALOG_IMAGE_FIELD_COPY.uploading
                : CATALOG_IMAGE_FIELD_COPY.chooseFile}
            </Button>
            <span className="catalog-image-field-status">
              {isUploading
                ? CATALOG_IMAGE_FIELD_COPY.uploading
                : CATALOG_IMAGE_FIELD_COPY.noFile}
            </span>
          </>
        )}
      </div>
    </div>
  );
}
