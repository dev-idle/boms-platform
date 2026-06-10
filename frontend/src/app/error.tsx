"use client";

import { useEffect } from "react";

import { Button } from "@/components/ui/button";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto flex max-w-lg flex-col gap-4 px-6 py-24 text-center">
      <h1 className="text-empty-title text-foreground">
        Something went wrong
      </h1>
      <p className="text-sm text-muted">
        {error.message || "An unexpected error occurred."}
      </p>
      <Button className="mx-auto" onClick={() => reset()} type="button">
        Try again
      </Button>
    </div>
  );
}
