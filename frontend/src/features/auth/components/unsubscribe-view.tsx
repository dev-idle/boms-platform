"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { ApiErrorCode, isApiError } from "@/lib/errors";

import { useLinkToken, useUnsubscribe } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
import { focusOnMount } from "../lib/focus-on-mount";
import { AuthFormShell } from "./auth-form-shell";
import { AuthFormSkeleton } from "./auth-form-skeleton";

/**
 * Stops promotion emails from the link in one of them as soon as the page
 * opens, without signing in. Asking again is harmless, so the page simply
 * says they are stopped.
 */
export function UnsubscribeView() {
  const token = useLinkToken();
  const unsubscribe = useUnsubscribe();
  const sent = useRef(false);

  useEffect(() => {
    if (!token || sent.current) {
      return;
    }
    sent.current = true;
    unsubscribe.mutate(token);
  }, [token, unsubscribe]);

  if (token === undefined || (token && (unsubscribe.isIdle || unsubscribe.isPending))) {
    return <AuthFormSkeleton {...AUTH_FORM_COPY.unsubscribe} fields={0} />;
  }

  // Each outcome replaces the skeleton, so its title takes the focus and is read out.
  if (unsubscribe.isSuccess) {
    return (
      <AuthFormShell
        description="You will no longer get promotion emails. Order emails continue. You can turn promotions back on in your account."
        title="You are unsubscribed"
        titleRef={focusOnMount}
      >
        <Button asChild className="w-full">
          <Link href={ROUTE.products}>Continue shopping</Link>
        </Button>
      </AuthFormShell>
    );
  }

  const linkUnusable =
    token === null || (isApiError(unsubscribe.error) && unsubscribe.error.code === ApiErrorCode.InvalidLink);
  return (
    <AuthFormShell
      description={
        linkUnusable
          ? "This unsubscribe link is incomplete or no longer valid. You can turn promotions off in your account."
          : "We could not unsubscribe you just now. Please try the link again in a moment."
      }
      title={linkUnusable ? "This link does not work" : "Something went wrong"}
      titleRef={focusOnMount}
    >
      {token && !linkUnusable ? (
        <Button className="w-full" onClick={() => unsubscribe.mutate(token)} type="button">
          Try again
        </Button>
      ) : null}
      <Button asChild className="w-full" variant="outline">
        <Link href={ROUTE.customer.account.profile}>Go to your account</Link>
      </Button>
    </AuthFormShell>
  );
}
