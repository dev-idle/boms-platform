import { BRAND } from "@/constants/brand";
import { cn } from "@/lib/utils";

type BrandWordmarkProps = {
  className?: string;
};

export function BrandWordmark({ className }: BrandWordmarkProps) {
  return (
    // Face, weight and colour live in `.brand-wordmark` — Tailwind utilities sit
    // in a later layer and would override the display-face contract.
    <span aria-label={BRAND.name} className={cn("brand-wordmark", className)}>
      <span className="brand-wordmark-body">Chou</span>
      <span className="brand-wordmark-accent">x</span>
    </span>
  );
}
