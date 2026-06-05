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

export function formatPriceCents(cents: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(cents / 100);
}
