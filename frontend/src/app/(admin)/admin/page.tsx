export default function AdminDashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Dashboard</h1>
      <p className="mt-2 max-w-prose text-sm text-zinc-600 dark:text-zinc-400">
        Operational metrics should load from Fiber through the DAL with role checks enforced server-side
        (Fiber must validate <code className="font-mono">X-User-Role</code> against the signed session, not trust it
        blindly).
      </p>
    </div>
  );
}
