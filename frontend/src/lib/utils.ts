import { clsx, type ClassValue } from "clsx";
import { extendTailwindMerge } from "tailwind-merge";

/**
 * The type styles the component CSS defines as `text-*` classes (`@layer components`).
 * tailwind-merge reads any unknown `text-*` class as a color utility, so a real
 * color after one (`text-form-label text-error`) would drop the type style. Each
 * is its own class group instead: a utility next to it refines it, and it never
 * replaces, or is replaced by, anything but itself.
 */
const TYPE_STYLES = [
  "caption",
  "caption-dashboard",
  "empty-title",
  "form-input",
  "form-label",
  "order-code",
  "overline",
  "page-title",
  "price",
  "section-heading",
  "table-cell",
  "tabular",
] as const;

type TypeStyleGroup = `type-${(typeof TYPE_STYLES)[number]}`;

const twMerge = extendTailwindMerge<TypeStyleGroup>({
  extend: {
    classGroups: Object.fromEntries(
      TYPE_STYLES.map((style) => [`type-${style}`, [`text-${style}`]]),
    ),
  },
});

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
