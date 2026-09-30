"use client";

import { Button } from "@/components/ui/button";

import { useDownloadMyData } from "../hooks";

/** The right of access: a copy of everything the bakery holds about you, as a file. */
export function DownloadMyDataCard() {
  const download = useDownloadMyData();

  return (
    <div className="storefront-account-data">
      <p className="storefront-account-data__copy">
        Your account, profile, the policies you accepted, and every order with its items and
        history, in one JSON file.
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
