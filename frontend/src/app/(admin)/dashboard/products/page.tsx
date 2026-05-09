export default function AdminProductsPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Admin · Products</h1>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        CRUD surfaces stay RSC-first; mutations go through Server Actions calling the DAL.
      </p>
    </div>
  );
}
