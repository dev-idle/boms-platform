import Link from "next/link";

import { Button } from "@/components/ui/button";
import { AUTH_REFRESH_COOKIE } from "@/constants/cookies";
import { ROUTE } from "@/constants/routes";

export default function HomePage() {
  return (
    <div className="flex min-h-full flex-col items-center justify-center px-6 py-24">
      <main className="flex w-full max-w-2xl flex-col gap-10 rounded-lg border border-border bg-surface p-12">
        <div className="flex flex-col gap-4">
          <p className="text-sm font-medium uppercase tracking-wide text-subtle">
            Bakery Ordering and Management System
          </p>
          <h1 className="font-heading text-3xl font-medium tracking-tight text-foreground">
            BOMS storefront
          </h1>
          <p className="text-lg leading-relaxed text-muted">
            Access tokens live in memory only. Refresh sessions use the HttpOnly{" "}
            <code className="rounded-sm bg-surface-alt px-1 py-0.5 text-sm text-foreground">
              {AUTH_REFRESH_COOKIE}
            </code>{" "}
            cookie (site path{" "}
            <code className="rounded-sm bg-surface-alt px-1 py-0.5 text-sm text-foreground">/</code>
            , used only on{" "}
            <code className="rounded-sm bg-surface-alt px-1 py-0.5 text-sm text-foreground">
              /api/v1/auth/*
            </code>
            ).
          </p>
        </div>
        <div className="flex flex-wrap gap-3">
          <Button asChild>
            <Link href={ROUTE.products}>Browse products</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href={ROUTE.login}>Sign in</Link>
          </Button>
        </div>
        <p className="border-t border-border pt-8 text-xs text-subtle">
          Next.js 16 · React 19 · Tailwind v4
        </p>
      </main>
    </div>
  );
}
