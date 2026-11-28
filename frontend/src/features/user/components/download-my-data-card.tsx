"use client";

import { Button } from "@/components/ui/button";

import { useDownloadMyData } from "../hooks";

/** The right of access: a copy of everything the bakery holds about you, as a file. */
export function DownloadMyDataCard() {
  const download = useDownloadMyData();

  return (
    <div className="storefront-account-data">
      <p className="storefront-account-data__copy">
        Your account, profile, sign-in history, cart, saved lists, reviews, messages and every
        order with its payment and history, in one file.
      </p>
      <Button
        aria-busy={download.isPending || undefined}
        disabled={download.isPending}
        onClick={() => download.mutate()}
        type="button"
        variant="outline"
      >
        Download my data
      </Button>
    </div>
  );
}
