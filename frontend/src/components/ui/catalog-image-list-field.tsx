"use client";

import { useRef, useState } from "react";
import { toast } from "sonner";
import {
  DashboardDeleteIcon,
  DashboardPrimaryIcon,
  DashboardViewIcon,
} from "@/components/icons/dashboard-ui-icons";
import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CATALOG_IMAGE_FIELD_COPY } from "@/constants/dashboard-form-copy";
import { isApiError, ApiErrorCode } from "@/lib/errors";
import {
  cloudinaryUploadSuccessMessage,
  planCatalogImageUpload,
} from "@/lib/cloudinary/catalog-image-upload";
import {
  CLOUDINARY_ALLOWED_IMAGE_TYPES,
  CLOUDINARY_MAX_PRODUCT_IMAGES,
  catalogProductImageUrl,
  isCloudinaryConfigured,
} from "@/lib/cloudinary/config";
import {
  catalogFileLabelForUrl,
  cloudinaryDeliveryFileLabel,
  isDuplicateCatalogFileLabel,
} from "@/lib/cloudinary/format";
import {
  CLOUDINARY_UPLOAD_COPY,
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

function isPreviewableImageUrl(url: string): boolean {
  const trimmed = url.trim();
  return trimmed.startsWith("http://") || trimmed.startsWith("https://");
}

type CatalogImagePreviewContentProps = {
  label: string;
  url: string;
};

function CatalogImagePreviewContent({ label, url }: CatalogImagePreviewContentProps) {
  const [loadState, setLoadState] = useState<"loading" | "loaded" | "error">("loading");
  const previewSrc = catalogProductImageUrl(url, 1200);

  if (!previewSrc) {
    return (
      <div className="catalog-image-preview-frame catalog-image-preview-frame--error">
        <p className="catalog-image-preview-status catalog-image-preview-status--error">
          {CATALOG_IMAGE_FIELD_COPY.previewLoadError}
        </p>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "catalog-image-preview-frame",
        loadState === "loading" && "catalog-image-preview-frame--loading",
        loadState === "loaded" && "catalog-image-preview-frame--loaded",
        loadState === "error" && "catalog-image-preview-frame--error",
      )}
    >
      {loadState === "loading" ? (
        <p className="sr-only">{CATALOG_IMAGE_FIELD_COPY.previewLoading}</p>
      ) : null}
      {loadState === "error" ? (
        <p className="catalog-image-preview-status catalog-image-preview-status--error">
          {CATALOG_IMAGE_FIELD_COPY.previewLoadError}
        </p>
      ) : null}
      {/* eslint-disable-next-line @next/next/no-img-element -- manager preview of external Cloudinary URLs */}
      <img
        alt={label}
        className={cn(
          "catalog-image-preview",
          loadState === "loaded" && "catalog-image-preview--loaded",
        )}
        decoding="async"
        hidden={loadState === "error"}
        onError={() => setLoadState("error")}
        onLoad={() => setLoadState("loaded")}
        src={previewSrc}
      />
    </div>
  );
}

export function CatalogImageListField({
  disabled = false,
  maxImages = CLOUDINARY_MAX_PRODUCT_IMAGES,
  onChange,
  value,
}: CatalogImageListFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [uploadedLabels, setUploadedLabels] = useState<Record<string, string>>({});

  const isDisabled = disabled || isUploading;
  const canAddMore = value.length < maxImages;
  const useCloudinary = isCloudinaryConfigured();
  const showOrderBadges = value.length > 1;

  function labelForUrl(url: string): string {
    return catalogFileLabelForUrl(url, uploadedLabels);
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    event.target.value = "";
    if (files.length === 0 || !canAddMore) {
      return;
    }

    const plan = planCatalogImageUpload({
      existingUrls: value,
      files,
      labelsByUrl: uploadedLabels,
      maxImages,
    });
    if (!plan.ok) {
      if (plan.message) {
        toast.error(plan.message);
      }
      return;
    }

    setIsUploading(true);
    try {
      const signature = await fetchCloudinaryUploadSignature();
      const uploadedUrls: string[] = [];
      const nextLabels: Record<string, string> = {};

      for (const file of plan.files) {
        const secureUrl = await uploadImageToCloudinary(file, signature);
        if (value.includes(secureUrl) || uploadedUrls.includes(secureUrl)) {
          toast.error(CLOUDINARY_UPLOAD_COPY.duplicateImage);
          continue;
        }
        uploadedUrls.push(secureUrl);
        nextLabels[secureUrl] = file.name;
      }

      if (uploadedUrls.length === 0) {
        return;
      }

      setUploadedLabels((current) => ({ ...current, ...nextLabels }));
      onChange([...value, ...uploadedUrls]);
      toast.success(cloudinaryUploadSuccessMessage(uploadedUrls.length));
    } catch (error) {
      const message = isApiError(error)
        ? error.code === ApiErrorCode.ServiceUnavailable
          ? CLOUDINARY_UPLOAD_COPY.uploadServiceUnavailable
          : error.message
        : error instanceof Error && error.message
          ? error.message
          : CLOUDINARY_UPLOAD_COPY.uploadFailed;
      toast.error(message);
    } finally {
      setIsUploading(false);
    }
  }

  function handleSetPrimary(index: number): void {
    if (index <= 0 || index >= value.length) {
      return;
    }
    const next = [...value];
    const [primary] = next.splice(index, 1);
    if (!primary) {
      return;
    }
    next.unshift(primary);
    onChange(next);
  }

  function handleRemove(url: string): void {
    if (previewUrl === url) {
      setPreviewUrl(null);
    }
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
    if (
      isPreviewableImageUrl(nextUrl) &&
      isDuplicateCatalogFileLabel(
        cloudinaryDeliveryFileLabel(nextUrl),
        value.filter((_, itemIndex) => itemIndex !== index),
        uploadedLabels,
      )
    ) {
      toast.error(CLOUDINARY_UPLOAD_COPY.duplicateFileName);
      return;
    }
    onChange(next);
  }

  function openFilePicker(): void {
    inputRef.current?.click();
  }

  const previewLabel =
    previewUrl != null ? labelForUrl(previewUrl) : CATALOG_IMAGE_FIELD_COPY.view;

  return (
    <>
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
            multiple
            onChange={handleFileChange}
            ref={inputRef}
            type="file"
          />
        ) : null}

        {value.length > 0 ? (
          <ul
            aria-label={CATALOG_IMAGE_FIELD_COPY.imageListAriaLabel}
            className="catalog-image-list-items"
          >
            {value.map((url, index) => {
              const label = labelForUrl(url);
              const canPreview = isPreviewableImageUrl(url);
              const isPrimary = index === 0;
              const order = index + 1;

              return (
                <li
                  aria-label={CATALOG_IMAGE_FIELD_COPY.imagePosition(order, value.length)}
                  className={cn(
                    "catalog-image-list-item",
                    isPrimary && "catalog-image-list-item--primary",
                  )}
                  key={`${url}-${index}`}
                >
                  <div className="catalog-image-list-item-main">
                    <div
                      className={cn(
                        "catalog-image-list-thumb-wrap",
                        isPrimary && "catalog-image-list-thumb-wrap--primary",
                      )}
                    >
                      {canPreview ? (
                        // eslint-disable-next-line @next/next/no-img-element -- manager preview of external Cloudinary URLs
                        <img
                          alt=""
                          className="catalog-image-list-thumb"
                          decoding="async"
                          height={48}
                          loading="lazy"
                          src={catalogProductImageUrl(url, 112)}
                          width={48}
                        />
                      ) : (
                        <span
                          aria-hidden
                          className="catalog-image-list-thumb catalog-image-list-thumb--empty"
                        />
                      )}
                      {showOrderBadges ? (
                        <span
                          aria-hidden
                          className={cn(
                            "catalog-image-list-order",
                            isPrimary && "catalog-image-list-order--primary",
                          )}
                        >
                          {order}
                        </span>
                      ) : null}
                    </div>
                    <div className="catalog-image-list-copy">
                      {isPrimary ? (
                        <span className="catalog-image-list-primary-label">
                          {CATALOG_IMAGE_FIELD_COPY.primaryImage}
                        </span>
                      ) : null}
                      {useCloudinary ? (
                        <span className="catalog-image-field-name" title={label}>
                          {label}
                        </span>
                      ) : (
                        <Input
                          disabled={isDisabled}
                          onChange={(event) => handleUrlChange(index, event.target.value)}
                          placeholder="https://example.com/image.jpg"
                          value={url}
                        />
                      )}
                    </div>
                  </div>
                  <div
                    aria-label={CATALOG_IMAGE_FIELD_COPY.actionsAriaLabel}
                    className="catalog-image-field-actions"
                    role="group"
                  >
                    <span className="catalog-image-field-action-slot">
                      {!isPrimary ? (
                        <button
                          aria-label={CATALOG_IMAGE_FIELD_COPY.setPrimaryImageFor(label)}
                          className="db-table-action db-table-action--edit"
                          disabled={isDisabled}
                          onClick={() => handleSetPrimary(index)}
                          title={CATALOG_IMAGE_FIELD_COPY.setPrimaryImage}
                          type="button"
                        >
                          <DashboardPrimaryIcon className="db-table-action-icon" />
                        </button>
                      ) : null}
                    </span>
                    <button
                      aria-label={CATALOG_IMAGE_FIELD_COPY.viewImage(label)}
                      className="db-table-action"
                      disabled={isDisabled || !canPreview}
                      onClick={() => setPreviewUrl(url)}
                      title={CATALOG_IMAGE_FIELD_COPY.view}
                      type="button"
                    >
                      <DashboardViewIcon className="db-table-action-icon" />
                    </button>
                    <button
                      aria-label={CATALOG_IMAGE_FIELD_COPY.remove}
                      className="db-table-action db-table-action--delete"
                      disabled={isDisabled}
                      onClick={() => handleRemove(url)}
                      title={CATALOG_IMAGE_FIELD_COPY.remove}
                      type="button"
                    >
                      <DashboardDeleteIcon className="db-table-action-icon" />
                    </button>
                  </div>
                </li>
              );
            })}
            </ul>
        ) : null}

        {useCloudinary ? (
          <div className="catalog-image-field-row">
            <Button
              disabled={isDisabled || !canAddMore}
              onClick={openFilePicker}
              size="sm"
              type="button"
              variant="outline"
            >
              {CATALOG_IMAGE_FIELD_COPY.addImage}
            </Button>
            <span className="catalog-image-field-status">
              {isUploading
                ? CATALOG_IMAGE_FIELD_COPY.uploadProgress
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

      <AppDialog
        description={previewLabel}
        footer={
          <AppDialogFooterActions>
            <Button onClick={() => setPreviewUrl(null)} type="button" variant="outline">
              {CATALOG_IMAGE_FIELD_COPY.closePreview}
            </Button>
          </AppDialogFooterActions>
        }
        onClose={() => setPreviewUrl(null)}
        open={previewUrl !== null}
        panelClassName="app-dialog-panel--catalog-image-preview"
        size="lg"
        title={CATALOG_IMAGE_FIELD_COPY.viewImageDialogTitle}
      >
        {previewUrl && isPreviewableImageUrl(previewUrl) ? (
          <CatalogImagePreviewContent
            key={previewUrl}
            label={previewLabel}
            url={previewUrl}
          />
        ) : null}
      </AppDialog>
    </>
  );
}
