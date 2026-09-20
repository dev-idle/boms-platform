import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";

import { cn } from "@/lib/utils";

const inputVariants = cva("field-chrome text-form-input w-full", {
  variants: {
    variant: {
      default: "",
      /** Neutral chrome for toolbar / table filters (no mint fill). */
      inline: "field-chrome--inline",
    },
  },
  defaultVariants: {
    variant: "default",
  },
});

export type InputProps = React.ComponentProps<"input"> &
  VariantProps<typeof inputVariants>;

export function Input({ className, type, variant, readOnly, ...props }: InputProps) {
  return (
    <input
      type={type}
      className={cn(
        inputVariants({ variant }),
        readOnly && "field-chrome--readonly",
        className,
      )}
      readOnly={readOnly}
      {...props}
    />
  );
}
