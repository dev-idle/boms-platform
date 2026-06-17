type StorefrontPageHeaderProps = {
  lead?: string;
  title: string;
};

export function StorefrontPageHeader({ lead, title }: StorefrontPageHeaderProps) {
  return (
    <header className="storefront-page-header">
      <h1 className="storefront-page-title">{title}</h1>
      {lead ? <p className="storefront-page-lead">{lead}</p> : null}
    </header>
  );
}
