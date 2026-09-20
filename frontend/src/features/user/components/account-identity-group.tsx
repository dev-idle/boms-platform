import { ReadonlyFieldGroup, type ReadonlyFieldRow } from "@/components/ui/readonly-field-group";

type AccountIdentityGroupProps =
  | { email: string; variant: "customer" | "admin" }
  | { email: string; employeeCode: string; variant: "operational" };

/**
 * Readonly identity on self-service profiles. The heading is chosen by who locked
 * the value, so the copy never points at the reader.
 */
export function AccountIdentityGroup(props: AccountIdentityGroupProps) {
  const rows: ReadonlyFieldRow[] = [{ label: "Email", value: props.email }];

  if (props.variant === "operational") {
    rows.push({ label: "Employee code", mono: true, value: props.employeeCode });

    return (
      <ReadonlyFieldGroup
        heading="Managed by an admin"
        reason={
          <>
            These identify you on every order you touch, so an admin changes them from{" "}
            <strong>Users</strong> — not here.
          </>
        }
        rows={rows}
      />
    );
  }

  return (
    <ReadonlyFieldGroup
      heading="Sign-in identity"
      reason={
        props.variant === "customer"
          ? "Your email is how you sign in. Contact us if you need it changed."
          : "Your email is how you sign in. Changing it is a support operation, not a profile edit."
      }
      rows={rows}
    />
  );
}
