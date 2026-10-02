"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { ApiErrorCode, isApiError } from "@/lib/errors";

import { useLinkToken, useUnsubscribe } from "../hooks";
import { AUTH_FORM_COPY } from "../lib/auth-form-copy";
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

  if (unsubscribe.isSuccess) {
    return (
      <AuthFormShell
        description="We will email you no more promotions. Emails about your orders still come. Changed your mind? Turn them back on from your account."
        title="You are unsubscribed"
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
          ? "This link is not one of ours. You can turn promotions off from your account instead."
          : "We could not unsubscribe you just now. Please try the link again in a moment."
      }
      title={linkUnusable ? "This link does not work" : "Something went wrong"}
    >
      <Button asChild className="w-full" variant="outline">
        <Link href={ROUTE.customer.account.profile}>Go to your account</Link>
      </Button>
    </AuthFormShell>
  );
}
