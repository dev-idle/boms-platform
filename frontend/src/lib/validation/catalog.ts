import { z } from "zod";

export const catalogSlugSchema = z
  .string()
  .trim()
  .min(1, "Slug is required")
  .max(128)
  .regex(
    /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
    "Slug must be lowercase letters, numbers, and hyphens",
  );

/**
 * Derive a URL-safe catalog slug from a display name (create-mode default).
 * Keep in sync with backend catalog.SlugFromName (NFD diacritic fold).
 */
export function slugifyCatalogName(name: string): string {
  const normalized = name
    .trim()
    .toLowerCase()
    .normalize("NFD")
    .replace(/\p{M}/gu, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 128);

  return normalized;
}

export function formatPriceCents(cents: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(cents / 100);
}
