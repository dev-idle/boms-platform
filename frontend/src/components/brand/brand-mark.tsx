import { cva, type VariantProps } from "class-variance-authority";

import {
  BRAND_MARK_PATH_PARTS,
  BRAND_MARK_PATHS,
  BRAND_MARK_STROKE_WIDTH,
  BRAND_MARK_VIEW_BOX,
} from "@/constants/brand-mark";
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

/** Matcha tree line-art (Choux Matcha v3). Geometry: `constants/brand-mark.ts`. */
export function BrandMark({ className, size }: BrandMarkProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn(brandMarkVariants({ size }), "brand-mark", className)}
      fill="none"
      shapeRendering="geometricPrecision"
      viewBox={`0 0 ${BRAND_MARK_VIEW_BOX} ${BRAND_MARK_VIEW_BOX}`}
      xmlns="http://www.w3.org/2000/svg"
    >
      {BRAND_MARK_PATHS.map((path, index) => (
        <path
          key={BRAND_MARK_PATH_PARTS[index]}
          className={cn(
            "brand-mark-part stroke-matcha-500",
            `brand-mark-${BRAND_MARK_PATH_PARTS[index]}`,
          )}
          d={path}
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={BRAND_MARK_STROKE_WIDTH}
        />
      ))}
    </svg>
  );
}
