export default function AdminUsersPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Admin · Users</h1>
      <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
        Never pass raw user records to Client Components without tainting sensitive fields first.
      </p>
    </div>
  );
}
