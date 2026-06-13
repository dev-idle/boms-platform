type DashboardActivityFeedPagePlaceholdersProps = {
  count: number;
};

/** Invisible feed items that pad a partial page to a full page height. */
export function DashboardActivityFeedPagePlaceholders({
  count,
}: DashboardActivityFeedPagePlaceholdersProps) {
  if (count <= 0) {
    return null;
  }

  return Array.from({ length: count }, (_, index) => (
    <li
      key={`activity-page-placeholder-${index}`}
      aria-hidden
      className="dashboard-activity-feed-item dashboard-activity-feed-item--page-placeholder"
    >
      <div className="dashboard-activity-feed-head">
        <p className="dashboard-activity-feed-summary">{"\u00a0"}</p>
        <time className="dashboard-activity-feed-when">{"\u00a0"}</time>
      </div>
      <p className="dashboard-activity-feed-meta">
        <span>{"\u00a0"}</span>
      </p>
    </li>
  ));
}
