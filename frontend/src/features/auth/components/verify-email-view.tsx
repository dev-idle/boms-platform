"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import { useLinkToken, useVerifyEmail } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { AuthFormShell } from "./auth-form-shell";
import { AuthFormSkeleton } from "./auth-form-skeleton";

/**
 * Confirms an address from the emailed link as soon as the page opens. It needs
 * no sign-in: the link is often opened on another device.
 */
export function VerifyEmailView() {
  const token = useLinkToken();
  const verify = useVerifyEmail();
  const signedIn = useAuthStore((state) => state.status === "authenticated");
  const confirmed = useAuthStore((state) => state.user?.email_verified === true);
  // One confirmation per link: a second effect run (Strict Mode) must not send
  // the used token again.
  const sent = useRef(false);

  useEffect(() => {
    if (!token || sent.current) {
      return;
    }
    sent.current = true;
    verify.mutate(token);
  }, [token, verify]);

  if (token === undefined || (token && (verify.isIdle || verify.isPending))) {
    return <AuthFormSkeleton {...AUTH_FORM_COPY.verifyEmail} fields={0} />;
  }

  const onward = signedIn
    ? { href: ROUTE.products, label: "Continue shopping" }
    : { href: ROUTE.login, label: "Sign in" };

  if (verify.isSuccess) {
    return (
      <AuthFormShell description="You can now place orders, and updates about them come to this address." title="Your email is confirmed">
        <Button asChild className="w-full">
          <Link href={onward.href}>{onward.label}</Link>
        </Button>
      </AuthFormShell>
    );
  }

  const linkUnusable = token === null || (isApiError(verify.error) && verify.error.code === ApiErrorCode.InvalidLink);
  if (linkUnusable && signedIn && confirmed) {
    return (
      <AuthFormShell description="This link was already used, and your address is confirmed: there is nothing more to do." title="Your email is already confirmed">
        <Button asChild className="w-full">
          <Link href={ROUTE.products}>Continue shopping</Link>
        </Button>
      </AuthFormShell>
    );
  }

  return (
    <AuthFormShell
      description={
        linkUnusable
          ? `This link has expired or has already been used. ${
              signedIn ? "Ask for a new one from your account page." : "Sign in to ask for a new one."
            }`
          : "We could not confirm your address just now. Please try the link again in a moment."
      }
      title={linkUnusable ? "This link no longer works" : "Something went wrong"}
    >
      <Button asChild className="w-full" variant="outline">
        <Link href={signedIn ? ROUTE.customer.account.profile : ROUTE.login}>
          {signedIn ? "Go to your account" : "Sign in"}
        </Link>
      </Button>
    </AuthFormShell>
  );
}
