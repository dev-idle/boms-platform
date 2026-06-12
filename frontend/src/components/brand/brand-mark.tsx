import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const brandMarkVariants = cva("shrink-0", {
  variants: {
    size: {
      sm: "size-7",
      md: "size-8",
      lg: "size-10",
      xl: "size-11",
      "2xl": "size-14",
    },
  },
  defaultVariants: {
    size: "sm",
  },
});

type BrandMarkProps = VariantProps<typeof brandMarkVariants> & {
  className?: string;
};

/** Minimal choux puff — body circle + piped cream cap (1.5 stroke) + one gold dot. */
export function BrandMark({ className, size }: BrandMarkProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn(brandMarkVariants({ size }), "brand-mark", className)}
      fill="none"
      shapeRendering="geometricPrecision"
      viewBox="0 0 32 32"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle
        className="brand-mark-part brand-mark-body stroke-matcha-500"
        cx="16"
        cy="19"
        r="8.25"
        strokeWidth="1.5"
      />
      <path
        className="brand-mark-part brand-mark-cream stroke-matcha-500"
        d="M9.4 14.6C10.2 10.8 12.8 8.6 16 8.6C19.2 8.6 21.8 10.8 22.6 14.6"
        strokeLinecap="round"
        strokeWidth="1.5"
      />
      <path
        className="brand-mark-part brand-mark-cream stroke-matcha-500"
        d="M16 8.6C15.2 7.4 15.4 6.2 16.4 5.3"
        strokeLinecap="round"
        strokeWidth="1.5"
      />
      <circle
        className="brand-mark-part brand-mark-dot fill-gold-500"
        cx="21.8"
        cy="6.6"
        r="1.5"
      />
    </svg>
  );
}
