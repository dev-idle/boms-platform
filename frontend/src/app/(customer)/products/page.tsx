export default function ProductsPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Products</h1>
      <p className="mt-2 max-w-prose text-sm text-zinc-600 dark:text-zinc-400">
        Fetch catalog data exclusively through <code className="font-mono">src/lib/dal</code> from a
        Server Component or Server Action. Validate every Fiber payload with Zod before use.
      </p>
    </div>
  );
}
