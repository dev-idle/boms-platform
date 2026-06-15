"use client";

import { useRef, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CATALOG_IMAGE_FIELD_COPY } from "@/constants/dashboard-form-copy";
import { isApiError } from "@/lib/errors";
import {
  CLOUDINARY_ALLOWED_IMAGE_TYPES,
  CLOUDINARY_MAX_IMAGE_BYTES,
  CLOUDINARY_MAX_PRODUCT_IMAGES,
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

type CatalogImageListFieldProps = {
  disabled?: boolean;
  maxImages?: number;
  onChange: (next: string[]) => void;
  value: string[];
};

export function CatalogImageListField({
  disabled = false,
  maxImages = CLOUDINARY_MAX_PRODUCT_IMAGES,
  onChange,
  value,
}: CatalogImageListFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadedLabels, setUploadedLabels] = useState<Record<string, string>>({});

  const isDisabled = disabled || isUploading;
  const canAddMore = value.length < maxImages;
  const useCloudinary = isCloudinaryConfigured();

  function labelForUrl(url: string): string {
    return uploadedLabels[url] ?? cloudinaryDeliveryFileLabel(url);
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || !canAddMore) {
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
      if (value.includes(secureUrl)) {
        toast.error(CLOUDINARY_UPLOAD_COPY.duplicateImage);
        return;
      }
      setUploadedLabels((current) => ({ ...current, [secureUrl]: file.name }));
      onChange([...value, secureUrl]);
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

  function handleRemove(url: string): void {
    setUploadedLabels((current) => {
      const next = { ...current };
      delete next[url];
      return next;
    });
    onChange(value.filter((item) => item !== url));
  }

  function handleUrlChange(index: number, nextUrl: string): void {
    const next = [...value];
    next[index] = nextUrl;
    onChange(next);
  }

  function openFilePicker(): void {
    inputRef.current?.click();
  }

  return (
    <div
      className={cn(
        "catalog-image-list",
        isUploading && "catalog-image-list--uploading",
        isDisabled && "catalog-image-list--disabled",
      )}
    >
      {useCloudinary ? (
        <input
          accept={CLOUDINARY_ALLOWED_IMAGE_TYPES.join(",")}
          aria-label={CATALOG_IMAGE_FIELD_COPY.uploadAriaLabel}
          className="sr-only"
          disabled={isDisabled || !canAddMore}
          onChange={handleFileChange}
          ref={inputRef}
          type="file"
        />
      ) : null}

      {value.length > 0 ? (
        <ul aria-label={CATALOG_IMAGE_FIELD_COPY.imageListAriaLabel} className="catalog-image-list-items">
          {value.map((url, index) => (
            <li className="catalog-image-list-item" key={`${url}-${index}`}>
              {useCloudinary ? (
                <span
                  className="catalog-image-field-name"
                  title={labelForUrl(url)}
                >
                  {labelForUrl(url)}
                </span>
              ) : (
                <Input
                  disabled={isDisabled}
                  onChange={(event) => handleUrlChange(index, event.target.value)}
                  placeholder="https://example.com/image.jpg"
                  value={url}
                />
              )}
              <Button
                disabled={isDisabled}
                onClick={() => handleRemove(url)}
                size="sm"
                type="button"
                variant="ghost"
              >
                {CATALOG_IMAGE_FIELD_COPY.remove}
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <p className="catalog-image-field-status">{CATALOG_IMAGE_FIELD_COPY.noImages}</p>
      )}

      {useCloudinary ? (
        <div className="catalog-image-field-row">
          <Button
            disabled={isDisabled || !canAddMore}
            onClick={openFilePicker}
            size="sm"
            type="button"
            variant="outline"
          >
            {isUploading
              ? CATALOG_IMAGE_FIELD_COPY.uploading
              : CATALOG_IMAGE_FIELD_COPY.addImage}
          </Button>
          <span className="catalog-image-field-status">
            {isUploading
              ? CATALOG_IMAGE_FIELD_COPY.uploading
              : CATALOG_IMAGE_FIELD_COPY.imageCount(value.length, maxImages)}
          </span>
        </div>
      ) : (
        <Button
          disabled={isDisabled || !canAddMore}
          onClick={() => onChange([...value, ""])}
          size="sm"
          type="button"
          variant="outline"
        >
          {CATALOG_IMAGE_FIELD_COPY.addImageUrl}
        </Button>
      )}
    </div>
  );
}
