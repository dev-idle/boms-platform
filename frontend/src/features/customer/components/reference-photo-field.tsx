"use client";

import { useRef, type ComponentProps } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { CLOUDINARY_ALLOWED_IMAGE_TYPES } from "@/lib/cloudinary/config";
import { CLOUDINARY_UPLOAD_COPY } from "@/lib/cloudinary/messages";
import { fetchReferenceUploadSignature } from "@/lib/cloudinary/signature";
import { uploadImageToCloudinary } from "@/lib/cloudinary/upload";
import { ApiErrorCode, isApiError } from "@/lib/errors";

/** The field's label, hint and error name the button that picks the file. */
type ReferencePhotoFieldProps = Pick<ComponentProps<"button">, "id" | "aria-describedby" | "aria-invalid"> & {
  /** The uploaded photo's URL; empty while there is none. */
  photo: string;
  disabled: boolean;
  uploading: boolean;
  onUploadingChange: (uploading: boolean) => void;
  onChange: (photo: string) => void;
};

function uploadErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    if (error.code === ApiErrorCode.EmailNotVerified) {
      return "Confirm your email address to upload a photo.";
    }
    if (error.code === ApiErrorCode.ServiceUnavailable) {
      return "Photo upload is not available right now. Describe the cake in the message instead.";
    }
  }
  return error instanceof Error ? error.message : CLOUDINARY_UPLOAD_COPY.uploadFailed;
}

/** A reference photo for a custom cake: uploaded straight to the bakery's image store, previewed, replaced or removed. */
export function ReferencePhotoField({
  id,
  "aria-describedby": describedBy,
  "aria-invalid": invalid,
  photo,
  disabled,
  uploading,
  onUploadingChange,
  onChange,
}: ReferencePhotoFieldProps) {
  const fileInput = useRef<HTMLInputElement>(null);

  async function upload(file: File): Promise<void> {
    onUploadingChange(true);
    try {
      onChange(await uploadImageToCloudinary(file, await fetchReferenceUploadSignature()));
    } catch (error) {
      toast.error(uploadErrorMessage(error));
    } finally {
      onUploadingChange(false);
    }
  }

  return (
    <div className="storefront-custom-cake__photo">
      {photo ? (
        // eslint-disable-next-line @next/next/no-img-element -- the customer's own Cloudinary upload
        <img alt="Your reference photo" className="storefront-custom-cake__preview" src={photo} />
      ) : null}
      <input
        ref={fileInput}
        accept={CLOUDINARY_ALLOWED_IMAGE_TYPES.join(",")}
        aria-label="Reference photo file"
        className="sr-only"
        disabled={disabled}
        tabIndex={-1}
        type="file"
        onChange={(event) => {
          const file = event.target.files?.[0];
          event.target.value = "";
          if (file) {
            void upload(file);
          }
        }}
      />
      <Button
        id={id}
        aria-busy={uploading || undefined}
        aria-describedby={describedBy}
        aria-invalid={invalid}
        disabled={disabled}
        type="button"
        variant="outline"
        onClick={() => fileInput.current?.click()}
      >
        {uploading ? "Uploading…" : photo ? "Replace photo" : "Upload a photo"}
      </Button>
      {photo ? (
        <Button disabled={disabled} type="button" variant="ghost" onClick={() => onChange("")}>
          Remove photo
        </Button>
      ) : null}
    </div>
  );
}
