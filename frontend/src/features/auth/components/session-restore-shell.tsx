/** Shared loading UI while auth store hydrates or session is restored. */
export function SessionRestoreShell() {
  return (
    <div
      className="flex min-h-[40vh] flex-1 items-center justify-center text-sm text-zinc-500"
      aria-live="polite"
      aria-busy="true"
    >
      Restoring session…
    </div>
  );
}
