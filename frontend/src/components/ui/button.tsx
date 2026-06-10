import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex min-h-11 items-center justify-center gap-2 whitespace-nowrap rounded-full px-6 text-sm font-medium transition-[background-color,box-shadow,transform,color] duration-standard ease-default focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-rose-200 disabled:pointer-events-none",
  {
    variants: {
      variant: {
        default:
          "bg-rose-500 text-surface hover:bg-rose-600 active:scale-[0.98] active:bg-rose-700 disabled:bg-rose-100 disabled:text-muted",
        outline:
          "border border-rose-200 bg-surface text-rose-500 hover:bg-rose-100 active:scale-[0.98] disabled:border-border disabled:bg-surface disabled:text-muted",
        ghost:
          "min-h-11 rounded-full px-4 text-ink hover:bg-blush active:scale-[0.98] disabled:text-muted",
        destructive:
          "bg-error text-surface hover:bg-error/90 active:scale-[0.98]",
      },
      size: {
        default: "h-11 px-6",
        sm: "h-9 min-h-9 px-4 text-sm",
        lg: "h-12 min-h-12 px-8 text-base",
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
  };

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  },
);
Button.displayName = "Button";
