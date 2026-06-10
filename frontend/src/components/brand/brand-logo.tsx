import { cva, type VariantProps } from "class-variance-authority";
import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { cn } from "@/lib/utils";

import { BrandMark } from "./brand-mark";
import { BrandWordmark } from "./brand-wordmark";

type BrandLogoSize = NonNullable<VariantProps<typeof brandLogoVariants>["size"]>;

const brandLogoVariants = cva(
  "inline-flex shrink-0 items-center leading-none text-ink",
  {
    variants: {
      size: {
        sm: "gap-2.5 text-[1.75rem]",
        md: "gap-3 text-[2.125rem]",
        header: "gap-2.5 text-[1.9rem] sm:gap-3 sm:text-[2.1rem] lg:text-[2.3rem]",
        nav: "gap-3 text-[2.25rem] sm:text-[2.5rem] lg:text-[2.65rem]",
        lg: "gap-3.5 text-[2.75rem] sm:text-3xl",
      },
    },
    defaultVariants: {
      size: "sm",
    },
  },
);

const markSizeForLogo: Record<
  BrandLogoSize,
  "sm" | "md" | "lg" | "xl" | "2xl"
> = {
  sm: "md",
  md: "lg",
  header: "lg",
  nav: "xl",
  lg: "2xl",
};

type BrandLogoProps = VariantProps<typeof brandLogoVariants> & {
  className?: string;
  /** When false, renders a non-interactive wordmark (e.g. footer heading). */
  linked?: boolean;
};

function BrandLogoInner({ size }: { size: BrandLogoSize | null | undefined }) {
  return (
    <>
      <BrandMark size={markSizeForLogo[size ?? "sm"]} />
      <BrandWordmark />
    </>
  );
}

export function BrandLogo({ className, size, linked = true }: BrandLogoProps) {
  if (!linked) {
    return (
      <span className={cn(brandLogoVariants({ size }), className)}>
        <BrandLogoInner size={size} />
      </span>
    );
  }

  return (
    <Link
      className={cn(brandLogoVariants({ size }), "brand-logo", className)}
      href={ROUTE.home}
    >
      <BrandLogoInner size={size} />
    </Link>
  );
}
