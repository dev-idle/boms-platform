import Link from "next/link";

import { ROUTE } from "@/constants/routes";

export default function LoginPage() {
  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-6 py-16">
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Sign in</h1>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Wire this route to Fiber authentication. Cookies must remain HttpOnly; never
        persist tokens in client storage.
      </p>
      <Link className="mt-8 text-sm font-medium text-zinc-900 underline dark:text-zinc-100" href={ROUTE.home}>
        Back to home
      </Link>
    </div>
  );
}
