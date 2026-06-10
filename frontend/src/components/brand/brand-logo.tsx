import { cva, type VariantProps } from "class-variance-authority";
import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { cn } from "@/lib/utils";

import { BrandMark } from "./brand-mark";

const brandWordmarkPrefix = BRAND.name.slice(0, -1);
const brandWordmarkAccent = BRAND.name.slice(-1);

const brandLogoVariants = cva(
  "inline-flex shrink-0 items-center font-heading font-normal tracking-tight text-ink",
  {
    variants: {
      size: {
        sm: "gap-2 text-lg",
        md: "gap-2.5 text-xl",
        lg: "gap-3 text-2xl",
      },
    },
    defaultVariants: {
      size: "sm",
    },
  },
);

type BrandLogoProps = VariantProps<typeof brandLogoVariants> & {
  className?: string;
  /** When false, renders a non-interactive wordmark (e.g. footer heading). */
  linked?: boolean;
};

function BrandLogoContent({
  className,
  size,
}: Pick<BrandLogoProps, "className" | "size">) {
  const resolvedSize = size ?? "sm";

  return (
    <span className={cn(brandLogoVariants({ size }), className)}>
      <BrandMark size={resolvedSize} />
      <span className="leading-none">
        {brandWordmarkPrefix}
        <span className="text-rose-500">{brandWordmarkAccent}</span>
      </span>
    </span>
  );
}

export function BrandLogo({
  className,
  size,
  linked = true,
}: BrandLogoProps) {
  const resolvedSize = size ?? "sm";

  if (!linked) {
    return <BrandLogoContent className={className} size={size} />;
  }

  return (
    <Link
      className={cn(
        brandLogoVariants({ size }),
        "group transition-colors duration-standard ease-default hover:text-rose-500",
        className,
      )}
      href={ROUTE.home}
    >
      <BrandMark
        className="transition-transform duration-standard ease-default group-hover:scale-105 motion-reduce:transform-none"
        size={resolvedSize}
      />
      <span className="leading-none">
        {brandWordmarkPrefix}
        <span className="text-rose-500 transition-colors duration-standard ease-default group-hover:text-rose-600">
          {brandWordmarkAccent}
        </span>
      </span>
    </Link>
  );
}
