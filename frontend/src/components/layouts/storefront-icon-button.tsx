import Link from "next/link";
import type { ReactNode, Ref } from "react";

import { cn } from "@/lib/utils";

type StorefrontIconButtonProps = {
  label: string;
  className?: string;
  children: ReactNode;
  href?: string;
  onClick?: () => void;
  pressed?: boolean;
  ref?: Ref<HTMLButtonElement>;
  "aria-controls"?: string;
};

export function StorefrontIconButton({
  label,
  className,
  children,
  href,
  onClick,
  pressed,
  ref,
  "aria-controls": ariaControls,
}: StorefrontIconButtonProps) {
  const classes = cn(
    "storefront-header-icon",
    pressed && "storefront-header-icon-pressed",
    className,
  );

  if (href) {
    return (
      <Link aria-label={label} className={classes} href={href} title={label}>
        {children}
      </Link>
    );
  }

  return (
    <button
      ref={ref}
      aria-controls={ariaControls}
      aria-label={label}
      aria-pressed={pressed}
      className={classes}
      onClick={onClick}
      title={label}
      type="button"
    >
      {children}
    </button>
  );
}
