import type { ReactNode } from "react";

import { FormControl, FormLabel } from "@/components/ui/form";

type AuthFormFieldControlProps = {
  label: string;
  children: ReactNode;
};

/** Label + input as one target — hover label highlights field; click label focuses input. */
export function AuthFormFieldControl({
  label,
  children,
}: AuthFormFieldControlProps) {
  return (
    <div className="auth-field-control">
      <FormLabel>{label}</FormLabel>
      <FormControl>{children}</FormControl>
    </div>
  );
}
