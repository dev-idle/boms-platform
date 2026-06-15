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

export function normalizeCatalogFileLabel(label: string): string {
  return label.trim().toLowerCase();
}

export function catalogFileLabelForUrl(
  url: string,
  labelsByUrl: Record<string, string>,
): string {
  return labelsByUrl[url] ?? cloudinaryDeliveryFileLabel(url);
}

export function isDuplicateCatalogFileLabel(
  fileLabel: string,
  urls: string[],
  labelsByUrl: Record<string, string>,
): boolean {
  const normalized = normalizeCatalogFileLabel(fileLabel);
  if (!normalized) {
    return false;
  }
  return urls.some(
    (url) =>
      normalizeCatalogFileLabel(catalogFileLabelForUrl(url, labelsByUrl)) ===
      normalized,
  );
}
