/** Cloudinary delivery + upload helpers (product catalog images). */

export const CLOUDINARY_DELIVERY_HOST = "res.cloudinary.com";

export const CLOUDINARY_DEFAULT_PRODUCT_UPLOAD_FOLDER = "boms/products";

export const CLOUDINARY_MAX_IMAGE_BYTES = 5 * 1024 * 1024; // Keep in sync with backend cloudinary.MaxProductImageBytes

/** Keep in sync with backend domain/product.MaxImagesPerProduct */
export const CLOUDINARY_MAX_PRODUCT_IMAGES = 5;

export const CLOUDINARY_ALLOWED_IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/avif",
] as const;

export function getCloudinaryCloudName(): string | undefined {
  const cloudName = process.env.NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME?.trim();
  return cloudName || undefined;
}

/** Must match CLOUDINARY_UPLOAD_FOLDER on the API when overriding the default. */
export function getCloudinaryUploadFolder(): string {
  const folder = process.env.NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER?.trim();
  if (!folder) {
    return CLOUDINARY_DEFAULT_PRODUCT_UPLOAD_FOLDER;
  }
  return folder.replace(/^\/+|\/+$/g, "");
}

export function isCloudinaryConfigured(): boolean {
  return Boolean(getCloudinaryCloudName());
}

function isCloudinaryVersionSegment(segment: string): boolean {
  if (segment.length < 2 || segment[0] !== "v") {
    return false;
  }
  return segment.slice(1).split("").every((char) => char >= "0" && char <= "9");
}

function isCloudinaryTransformSegment(segment: string): boolean {
  if (isCloudinaryVersionSegment(segment)) {
    return false;
  }
  return segment.includes(",") || segment.includes("_");
}

function normalizeCloudinaryPublicPath(publicPath: string): string | null {
  if (!publicPath || publicPath.includes("..")) {
    return null;
  }

  const segments = publicPath.split("/").filter((segment) => segment && segment !== ".");
  for (const segment of segments) {
    if (segment === "..") {
      return null;
    }
  }
  return segments.join("/");
}

function isPublicPathUnderFolder(publicPath: string, folder: string): boolean {
  const cleaned = normalizeCloudinaryPublicPath(publicPath);
  if (!cleaned) {
    return false;
  }
  return cleaned === folder || cleaned.startsWith(`${folder}/`);
}

function cloudinaryPublicIdPath(afterUpload: string): string {
  const segments = afterUpload.replace(/^\/+|\/+$/g, "").split("/");
  let start = 0;
  while (start < segments.length) {
    const segment = segments[start];
    if (!segment) {
      start += 1;
      continue;
    }
    if (isCloudinaryTransformSegment(segment) || isCloudinaryVersionSegment(segment)) {
      start += 1;
      continue;
    }
    break;
  }
  return segments.slice(start).join("/");
}

export function isCloudinaryDeliveryUrl(
  url: string,
  cloudName = getCloudinaryCloudName(),
): boolean {
  if (!cloudName) {
    return false;
  }
  try {
    const parsed = new URL(url);
    if (parsed.protocol !== "https:" || parsed.hostname !== CLOUDINARY_DELIVERY_HOST) {
      return false;
    }
    const prefix = `/${cloudName}/image/upload/`;
    return parsed.pathname.startsWith(prefix) && parsed.pathname.length > prefix.length;
  } catch {
    return false;
  }
}

export function isCloudinaryDeliveryUrlInFolder(
  url: string,
  folder = getCloudinaryUploadFolder(),
  cloudName = getCloudinaryCloudName(),
): boolean {
  if (!cloudName || !isCloudinaryDeliveryUrl(url, cloudName)) {
    return false;
  }

  const normalizedFolder = folder.replace(/^\/+|\/+$/g, "");
  if (!normalizedFolder) {
    return false;
  }

  try {
    const parsed = new URL(url);
    const prefix = `/${cloudName}/image/upload/`;
    if (!parsed.pathname.startsWith(prefix)) {
      return false;
    }
    const publicPath = cloudinaryPublicIdPath(parsed.pathname.slice(prefix.length));
    return isPublicPathUnderFolder(publicPath, normalizedFolder);
  } catch {
    return false;
  }
}

function isCloudinaryDeliveryUrlPattern(url: string): boolean {
  try {
    const parsed = new URL(url);
    if (parsed.protocol !== "https:" || parsed.hostname !== CLOUDINARY_DELIVERY_HOST) {
      return false;
    }
    const marker = "/image/upload/";
    const index = parsed.pathname.indexOf(marker);
    return index !== -1 && parsed.pathname.length > index + marker.length;
  } catch {
    return false;
  }
}

/** Storefront-friendly transforms for catalog cards and detail hero. */
export function catalogProductImageUrl(
  url: string | null | undefined,
  width = 800,
): string | undefined {
  if (!url) {
    return undefined;
  }
  if (!isCloudinaryDeliveryUrlPattern(url)) {
    return url;
  }
  const marker = "/image/upload/";
  const index = url.indexOf(marker);
  if (index === -1) {
    return url;
  }
  const prefix = url.slice(0, index + marker.length);
  const afterUpload = url.slice(index + marker.length);
  const publicIdPath = cloudinaryPublicIdPath(afterUpload);
  if (!publicIdPath) {
    return url;
  }
  return `${prefix}f_auto,q_auto,w_${width}/${publicIdPath}`;
}

export function primaryCatalogProductImageUrl(
  urls: string[] | null | undefined,
  width = 800,
): string | undefined {
  return catalogProductImageUrl(urls?.[0], width);
}
