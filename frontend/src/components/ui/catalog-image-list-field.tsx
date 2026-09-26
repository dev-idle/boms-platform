"use client";

import { Fragment, useRef, useState } from "react";
import { toast } from "sonner";
import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
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

type CatalogImageRowActionsProps = {
  /** Set while the field is busy — every action greys with this as its reason. */
  blockedReason?: string;
  canPreview: boolean;
  isPrimary: boolean;
  label: string;
  onPreview: () => void;
  onRemove: () => void;
  onSetPrimary: () => void;
  /** False when there is nothing to rank: one image, or a single-image field. */
  showPrimary: boolean;
};

/**
 * Row actions in words, not glyphs (02-COMPONENTS §"Row actions"): an icon in a
 * row needs a legend and this row has none, while one of them deletes.
 *
 * Ordered safe → destructive, the way the catalog tables beside this field order
 * Edit before Delete, so Remove is never the first label the eye lands on. Among
 * the safe ones the leading label is the one that changes the catalog — Primary
 * decides the card, the hero and the lead of the detail strip, while View only
 * looks. The hairline between members is what keeps a run of uppercase labels
 * readable as separate actions.
 *
 * `PRIMARY`, not `SET AS PRIMARY`: this row is the busiest in the dashboard, the
 * word matches the badge it produces, and the accessible name carries the verb
 * the label drops.
 */
function CatalogImageRowActions({
  blockedReason,
  canPreview,
  isPrimary,
  label,
  onPreview,
  onRemove,
  onSetPrimary,
  showPrimary,
}: CatalogImageRowActionsProps) {
  const actions = [
    showPrimary ? (
      <DashboardTableActionButton
        blockedReason={
          blockedReason ??
          (isPrimary ? CATALOG_IMAGE_FIELD_COPY.alreadyPrimary : undefined)
        }
        key="primary"
        label={CATALOG_IMAGE_FIELD_COPY.setPrimaryImageFor(label)}
        onClick={onSetPrimary}
        text={CATALOG_IMAGE_FIELD_COPY.primaryImage}
        tone="accent"
      />
    ) : null,
    <DashboardTableActionButton
      blockedReason={
        blockedReason ??
        (canPreview ? undefined : CATALOG_IMAGE_FIELD_COPY.previewUnavailable)
      }
      key="view"
      label={CATALOG_IMAGE_FIELD_COPY.viewImage(label)}
      onClick={onPreview}
      text={CATALOG_IMAGE_FIELD_COPY.view}
      tone="accent"
    />,
    <DashboardTableActionButton
      blockedReason={blockedReason}
      key="remove"
      label={CATALOG_IMAGE_FIELD_COPY.removeImage(label)}
      onClick={onRemove}
      text={CATALOG_IMAGE_FIELD_COPY.remove}
      tone="danger"
    />,
  ].filter(Boolean);

  return (
    <div
      aria-label={CATALOG_IMAGE_FIELD_COPY.actionsAriaLabel}
      className="dashboard-inline-actions"
      role="group"
    >
      {actions.map((action, index) => (
        <Fragment key={index}>
          {index > 0 ? (
            <span aria-hidden className="dashboard-inline-actions__sep" />
          ) : null}
          {action}
        </Fragment>
      ))}
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
  const dragDepth = useRef(0);
  const [isUploading, setIsUploading] = useState(false);
  const [isDropTarget, setIsDropTarget] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [uploadedLabels, setUploadedLabels] = useState<Record<string, string>>({});

  const isDisabled = disabled || isUploading;
  const canAddMore = value.length < maxImages;
  const useCloudinary = isCloudinaryConfigured();
  // A gallery ranks its images; a single-image field has nothing to rank, so the
  // order badge, the primary mark and the counter all belong to the gallery.
  const isGallery = maxImages > 1;
  const showOrderBadges = isGallery && value.length > 1;

  function labelForUrl(url: string): string {
    return catalogFileLabelForUrl(url, uploadedLabels);
  }

  async function uploadFiles(files: File[]): Promise<void> {
    // A drop and a paste reach this without passing the file input, which is
    // what `disabled` guards — so the guard lives here, not on the control. The
    // count is left to the planner below: it is the one that can say how many
    // still fit, and a drop on a full gallery deserves that message rather than
    // silence.
    if (files.length === 0 || isDisabled) {
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

  function handleFileChange(event: React.ChangeEvent<HTMLInputElement>): void {
    const files = Array.from(event.target.files ?? []);
    event.target.value = "";
    void uploadFiles(files);
  }

  /** A drag carrying anything but files is somebody else's — leave it alone. */
  function dragCarriesFiles(event: React.DragEvent): boolean {
    return Array.from(event.dataTransfer.types).includes("Files");
  }

  function handleDragEnter(event: React.DragEvent): void {
    if (!useCloudinary || !dragCarriesFiles(event)) {
      return;
    }
    event.preventDefault();
    // Entering a child fires enter before the parent's leave, so count depth
    // instead of toggling — otherwise the outline blinks across every row.
    dragDepth.current += 1;
    setIsDropTarget(true);
  }

  function handleDragOver(event: React.DragEvent): void {
    if (!useCloudinary || !dragCarriesFiles(event)) {
      return;
    }
    event.preventDefault();
    event.dataTransfer.dropEffect =
      isDisabled || !canAddMore ? "none" : "copy";
  }

  function handleDragLeave(event: React.DragEvent): void {
    if (!useCloudinary || !dragCarriesFiles(event)) {
      return;
    }
    dragDepth.current = Math.max(dragDepth.current - 1, 0);
    if (dragDepth.current === 0) {
      setIsDropTarget(false);
    }
  }

  /**
   * A drag dropped outside the window, or abandoned with Escape, never sends the
   * leave that would close the count — so the end of the drag itself resets it,
   * rather than leaving the field outlined until the next one.
   */
  function resetDragState(): void {
    dragDepth.current = 0;
    setIsDropTarget(false);
  }

  function handleDrop(event: React.DragEvent): void {
    if (!useCloudinary || !dragCarriesFiles(event)) {
      return;
    }
    event.preventDefault();
    dragDepth.current = 0;
    setIsDropTarget(false);
    void uploadFiles(Array.from(event.dataTransfer.files));
  }

  /** Scoped to the field, so Ctrl+V anywhere else in the form is untouched. */
  function handlePaste(event: React.ClipboardEvent): void {
    if (!useCloudinary) {
      return;
    }
    const files = Array.from(event.clipboardData.files);
    if (files.length === 0) {
      return;
    }
    event.preventDefault();
    void uploadFiles(files);
  }

  /** To the front — the first image is the card, the hero and the strip's lead. */
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
          isDropTarget && "catalog-image-list--drop",
        )}
        onDragEnd={resetDragState}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onPaste={handlePaste}
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
                  aria-label={CATALOG_IMAGE_FIELD_COPY.imagePosition(
                    order,
                    value.length,
                    showOrderBadges && isPrimary,
                  )}
                  className="catalog-image-list-item"
                  key={url}
                >
                  <div className="catalog-image-list-item-main">
                    <div
                      className={cn(
                        "catalog-image-list-thumb-wrap",
                        showOrderBadges &&
                          isPrimary &&
                          "catalog-image-list-thumb-wrap--primary",
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
                  <CatalogImageRowActions
                    blockedReason={
                      isDisabled ? CATALOG_IMAGE_FIELD_COPY.busy : undefined
                    }
                    canPreview={canPreview}
                    isPrimary={isPrimary}
                    label={label}
                    onPreview={() => setPreviewUrl(url)}
                    onRemove={() => handleRemove(url)}
                    onSetPrimary={() => handleSetPrimary(index)}
                    showPrimary={isGallery}
                  />
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
              {isGallery
                ? CATALOG_IMAGE_FIELD_COPY.addImage
                : CATALOG_IMAGE_FIELD_COPY.addOneImage}
            </Button>
            {isUploading || isGallery ? (
              <span className="catalog-image-field-status">
                {isUploading
                  ? CATALOG_IMAGE_FIELD_COPY.uploadProgress
                  : CATALOG_IMAGE_FIELD_COPY.imageCount(value.length, maxImages)}
              </span>
            ) : null}
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
