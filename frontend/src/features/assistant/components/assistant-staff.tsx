"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import { GeneralMessages } from "@/features/customer";
import { loginHrefPreservingNext } from "@/lib/validate-next";

type AssistantStaffProps = {
  signedIn: boolean;
};

/**
 * The customer's conversation with staff. Staff answer from an account, so a
 * guest signs in first, and comes back to the page they were on.
 */
export function AssistantStaff({ signedIn }: AssistantStaffProps) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const search = searchParams.toString() ? `?${searchParams.toString()}` : "";
  if (signedIn) {
    return <GeneralMessages />;
  }
  return (
    <div className="assistant-panel__guest">
      <p className="text-sm">Sign in to talk to our staff; their replies appear here.</p>
      <Button asChild size="sm">
        <Link href={loginHrefPreservingNext(pathname, search)}>Sign in</Link>
      </Button>
    </div>
  );
}
