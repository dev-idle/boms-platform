type StorefrontPageHeaderProps = {
  /** Micro-label above the title — the section this page belongs to. */
  eyebrow?: string;
  lead?: string;
  title: string;
};

export function StorefrontPageHeader({
  eyebrow,
  lead,
  title,
}: StorefrontPageHeaderProps) {
  return (
    <header className="storefront-page-header">
      <div className="storefront-page-header__main">
        {eyebrow ? <p className="text-overline">{eyebrow}</p> : null}
        <h1 className="storefront-page-title">{title}</h1>
      </div>
      {lead ? <p className="storefront-page-lead">{lead}</p> : null}
    </header>
  );
}
