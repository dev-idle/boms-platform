/** Human-readable size for upload validation messages (matches API max_bytes). */
export function formatCloudinaryMaxImageSize(bytes: number): string {
  const mib = bytes / (1024 * 1024);
  if (Number.isInteger(mib)) {
    return `${mib} MB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Best-effort filename label from a Cloudinary delivery URL (edit / reload). */
export function cloudinaryDeliveryFileLabel(url: string): string {
  try {
    const parsed = new URL(url);
    const segment = parsed.pathname.split("/").filter(Boolean).at(-1);
    if (segment) {
      return decodeURIComponent(segment);
    }
  } catch {
    // Fall through to generic label.
  }
  return "Product image";
}
