"use client";

import { useRouter } from "next/navigation";

import { StorefrontErrorState } from "@/components/layouts/storefront-state";

type HomeSectionErrorProps = {
  title: string;
};

/**
 * A home section the server could not read, under its header as the grid would
 * be: asking again re-reads the page's data.
 */
export function HomeSectionError({ title }: HomeSectionErrorProps) {
  const router = useRouter();

  return (
    <div className="mt-10">
      <StorefrontErrorState onRetry={() => router.refresh()} title={title} />
    </div>
  );
}
