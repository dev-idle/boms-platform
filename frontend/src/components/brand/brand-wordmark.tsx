import { BRAND } from "@/constants/brand";
import { cn } from "@/lib/utils";

type BrandWordmarkProps = {
  className?: string;
};

export function BrandWordmark({ className }: BrandWordmarkProps) {
  return (
    <span
      aria-label={BRAND.name}
      className={cn(
        "brand-wordmark font-display text-matcha-500 italic font-semibold",
        className,
      )}
    >
      <span className="brand-wordmark-body">Chou</span>
      <span className="brand-wordmark-accent">x</span>
    </span>
  );
}
