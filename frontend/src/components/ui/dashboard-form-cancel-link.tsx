import Link from "next/link";

import { Button } from "./button";

type DashboardFormCancelLinkProps = {
  /** The list the form page was opened from. */
  href: string;
};

/** Leaves a form page without saving, beside its primary action. */
export function DashboardFormCancelLink({ href }: DashboardFormCancelLinkProps) {
  return (
    <Button asChild variant="ghost">
      <Link href={href}>Cancel</Link>
    </Button>
  );
}
