import { BRAND } from "@/constants/brand";
import { cn } from "@/lib/utils";

type BrandWordmarkProps = {
  className?: string;
};

export function BrandWordmark({ className }: BrandWordmarkProps) {
  return (
    <span
      aria-label={BRAND.name}
      className={cn("brand-wordmark", className)}
    >
      <span className="brand-wordmark-body">Ros</span>
      <span className="brand-wordmark-accent">é</span>
    </span>
  );
}
