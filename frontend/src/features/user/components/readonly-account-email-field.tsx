import { useId } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type ReadonlyAccountEmailFieldProps = {
  email: string;
};

/** Sign-in identity from GET /me — not editable via PATCH /me. */
export function ReadonlyAccountEmailField({
  email,
}: ReadonlyAccountEmailFieldProps) {
  const emailId = useId();

  return (
    <div className="field-control field-control--readonly">
      <Label htmlFor={emailId}>Email</Label>
      <Input
        autoComplete="email"
        id={emailId}
        readOnly
        tabIndex={-1}
        type="email"
        value={email}
      />
    </div>
  );
}
