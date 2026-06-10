import Link from "next/link";
import { forwardRef, type ReactNode } from "react";

import { cn } from "@/lib/utils";

type StorefrontIconButtonProps = {
  label: string;
  className?: string;
  children: ReactNode;
  href?: string;
  onClick?: () => void;
  pressed?: boolean;
  "aria-controls"?: string;
};

export const StorefrontIconButton = forwardRef<
  HTMLButtonElement,
  StorefrontIconButtonProps
>(function StorefrontIconButton(
  {
    label,
    className,
    children,
    href,
    onClick,
    pressed,
    "aria-controls": ariaControls,
  },
  ref,
) {
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
});
