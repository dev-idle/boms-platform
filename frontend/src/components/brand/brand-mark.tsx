import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const brandMarkVariants = cva("shrink-0", {
  variants: {
    size: {
      sm: "size-5",
      md: "size-6",
      lg: "size-8",
      xl: "size-10",
    },
  },
  defaultVariants: {
    size: "sm",
  },
});

type BrandMarkProps = VariantProps<typeof brandMarkVariants> & {
  className?: string;
};

/**
 * Stylized rose bloom — three soft petals with a champagne center.
 * Scales cleanly from favicon to hero panels.
 */
export function BrandMark({ className, size }: BrandMarkProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn(brandMarkVariants({ size }), className)}
      fill="none"
      viewBox="0 0 40 40"
      xmlns="http://www.w3.org/2000/svg"
    >
      <g className="text-rose-500">
        <g opacity="0.45" transform="rotate(0 20 20)">
          <path
            d="M20 6C14.5 8 12.5 14 14.5 19.5C15.5 22.5 17.8 24.2 20 23.5C22.2 24.2 24.5 22.5 25.5 19.5C27.5 14 25.5 8 20 6Z"
            fill="currentColor"
          />
        </g>
        <g opacity="0.7" transform="rotate(120 20 20)">
          <path
            d="M20 6C14.5 8 12.5 14 14.5 19.5C15.5 22.5 17.8 24.2 20 23.5C22.2 24.2 24.5 22.5 25.5 19.5C27.5 14 25.5 8 20 6Z"
            fill="currentColor"
          />
        </g>
        <g transform="rotate(240 20 20)">
          <path
            d="M20 6C14.5 8 12.5 14 14.5 19.5C15.5 22.5 17.8 24.2 20 23.5C22.2 24.2 24.5 22.5 25.5 19.5C27.5 14 25.5 8 20 6Z"
            fill="currentColor"
          />
        </g>
      </g>
      <circle className="fill-gold-500" cx="20" cy="18.5" r="2.25" />
      <path
        className="stroke-gold-500"
        d="M20 24.5V31"
        opacity="0.55"
        strokeLinecap="round"
        strokeWidth="1.25"
      />
      <path
        className="stroke-gold-500"
        d="M20 28.5C17.5 29.5 16 31 16 31"
        opacity="0.4"
        strokeLinecap="round"
        strokeWidth="1"
      />
    </svg>
  );
}
