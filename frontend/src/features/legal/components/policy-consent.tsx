import Link from "next/link";
import { useId } from "react";

import { CheckboxField } from "@/components/ui/checkbox-field";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

type PolicyConsentProps = {
  checked: boolean;
  /** Id of the message that explains why the box must be ticked. */
  describedBy?: string;
  disabled?: boolean;
  invalid?: boolean;
  onCheckedChange: (checked: boolean) => void;
  /** Signing up agrees to the terms and privacy policy; an order to the terms and refund policy. */
  scope: "account" | "order";
};

/** The tick a customer gives the current policies, with links that open them beside the form. */
export function PolicyConsent({
  checked,
  describedBy,
  disabled,
  invalid,
  onCheckedChange,
  scope,
}: PolicyConsentProps) {
  const newTabHintId = useId();
  const other =
    scope === "account"
      ? { href: ROUTE.privacy, label: PAGE_TITLES.privacy }
      : { href: ROUTE.refundPolicy, label: PAGE_TITLES.refundPolicy };

  return (
    <>
      <CheckboxField
        aria-describedby={describedBy}
        aria-invalid={invalid || undefined}
        checked={checked}
        disabled={disabled}
        onChange={(event) => onCheckedChange(event.target.checked)}
      >
        I agree to the <PolicyLink describedBy={newTabHintId} href={ROUTE.terms} label={PAGE_TITLES.terms} /> and
        the <PolicyLink describedBy={newTabHintId} href={other.href} label={other.label} />
        {scope === "order" ? " for this order." : "."}
      </CheckboxField>
      <span hidden id={newTabHintId}>
        Opens in a new tab
      </span>
    </>
  );
}

function PolicyLink({ describedBy, href, label }: { describedBy: string; href: string; label: string }) {
  return (
    <Link aria-describedby={describedBy} href={href} rel="noopener" target="_blank">
      {label}
    </Link>
  );
}
