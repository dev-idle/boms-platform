import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/lib/utils";

/** Arrow-right for primary CTAs — minimal stroke. */
export function ButtonArrowIcon({ className }: { className?: string }) {
  return (
    <svg
      aria-hidden
      className={cn("btn-arrow size-[14px] shrink-0", className)}
      fill="none"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5 12h14m-7-7 7 7-7 7"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.75}
      />
    </svg>
  );
}

const buttonVariants = cva(
  [
    "btn-chrome inline-flex items-center justify-center gap-1.5 whitespace-nowrap",
    "rounded-button font-body text-sm font-medium leading-none tracking-[0.012em] antialiased",
  ].join(" "),
  {
    variants: {
      variant: {
        default: "btn-primary",
        outline: "btn-variant-outline",
        gold: "btn-variant-gold",
        ghost: "btn-variant-ghost",
        destructive: "btn-variant-destructive",
      },
      size: {
        default: "h-10 min-h-11 min-w-11 px-5",
        sm: "h-9 min-h-11 min-w-11 px-4 text-xs",
        lg: "h-11 min-h-11 min-w-11 px-6",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
    /** Primary CTA arrow (hero, add-to-cart). Secondary buttons omit this. */
    showArrow?: boolean;
  };

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { className, variant, size, asChild = false, showArrow = false, children, ...props },
    ref,
  ) => {
    const classes = cn(className, buttonVariants({ variant, size }));

    if (asChild) {
      if (showArrow) {
        const child = React.Children.only(children) as React.ReactElement<{
          children?: React.ReactNode;
        }>;

        return (
          <Slot className={classes} ref={ref} {...props}>
            {React.cloneElement(child, {
              children: (
                <>
                  {child.props.children}
                  <ButtonArrowIcon />
                </>
              ),
            })}
          </Slot>
        );
      }

      return (
        <Slot className={classes} ref={ref} {...props}>
          {children}
        </Slot>
      );
    }

    return (
      <button className={classes} ref={ref} {...props}>
        {children}
        {showArrow ? <ButtonArrowIcon /> : null}
      </button>
    );
  },
);
Button.displayName = "Button";

export { buttonVariants };
