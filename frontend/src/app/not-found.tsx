import Link from "next/link";

import { ROUTE } from "@/constants/routes";

export default function NotFound() {
  return (
    <div className="mx-auto flex max-w-lg flex-col gap-4 px-6 py-24 text-center">
      <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-50">Page not found</h1>
      <p className="text-sm text-zinc-600 dark:text-zinc-400">The requested resource does not exist.</p>
      <Link className="text-sm font-medium text-zinc-900 underline dark:text-zinc-100" href={ROUTE.home}>
        Return home
      </Link>
    </div>
  );
}
