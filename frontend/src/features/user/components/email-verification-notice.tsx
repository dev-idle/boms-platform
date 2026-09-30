"use client";

import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";

import { useMe, useResendVerificationEmail } from "../hooks";

type EmailVerificationNoticeProps = {
  /** What confirming unlocks where the notice shows. */
  reason: string;
};

/**
 * Tells a signed-in customer their address is not confirmed yet, and sends a
 * new link. Renders nothing once it is confirmed, or before the session is known.
 * The link is often opened on another device, so while the address is
 * unconfirmed the account is read again on mount and on focus, and the notice
 * goes away by itself.
 */
export function EmailVerificationNotice({ reason }: EmailVerificationNoticeProps) {
  const user = useAuthStore((state) => state.user);
  const resend = useResendVerificationEmail();
  useMe({ keepFresh: user?.email_verified === false });

  if (!user || user.email_verified) {
    return null;
  }

  return (
    <div className="email-verification-notice">
      <p className="email-verification-notice__label">Confirm your email</p>
      <p className="email-verification-notice__copy">
        {reason} We sent a link to <strong>{user.email}</strong>.
      </p>
      <Button
        aria-busy={resend.isPending || undefined}
        disabled={resend.isPending}
        type="button"
        variant="outline"
        onClick={() => resend.mutate()}
      >
        {resend.isPending ? "Sending…" : "Send a new link"}
      </Button>
    </div>
  );
}
