type DashboardTablePagePlaceholdersProps = {
  columnCount: number;
  count: number;
};

/** Invisible table rows that pad a partial page to a full page height. */
export function DashboardTablePagePlaceholders({
  columnCount,
  count,
}: DashboardTablePagePlaceholdersProps) {
  if (count <= 0) {
    return null;
  }

  return Array.from({ length: count }, (_, index) => (
    <tr
      key={`table-page-placeholder-${index}`}
      aria-hidden
      className="db-table-row--page-placeholder"
    >
      {Array.from({ length: columnCount }, (_, cellIndex) => (
        <td key={cellIndex}>{"\u00a0"}</td>
      ))}
    </tr>
  ));
}
