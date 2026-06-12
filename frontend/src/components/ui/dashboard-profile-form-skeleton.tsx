type DashboardProfileFormSkeletonProps = {
  fields?: number;
};

/** Placeholder while profile fields load inside a dashboard profile section. */
export function DashboardProfileFormSkeleton({
  fields = 2,
}: DashboardProfileFormSkeletonProps) {
  return (
    <div
      aria-busy="true"
      aria-label="Loading profile"
      className="dashboard-profile-form-skeleton"
      role="status"
    >
      {Array.from({ length: fields }, (_, index) => (
        <div className="dashboard-profile-form-skeleton-field" key={index}>
          <div className="skeleton dashboard-profile-form-skeleton-label" />
          <div className="skeleton dashboard-profile-form-skeleton-input" />
        </div>
      ))}
      <div className="skeleton dashboard-profile-form-skeleton-button" />
    </div>
  );
}
