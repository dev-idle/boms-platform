/** Non-interactive table cell for actions unavailable on the signed-in admin row. */
export function AdminTableActionUnavailable() {
  return (
    <span aria-hidden className="db-table-action-placeholder">
      —
    </span>
  );
}
