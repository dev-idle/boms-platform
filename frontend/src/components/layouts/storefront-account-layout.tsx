import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

type StorefrontAccountSectionProps = {
  children: ReactNode;
  description?: string;
  id: string;
  title: string;
  variant?: "default" | "danger";
};

/** Section block for unified customer account page. */
export function StorefrontAccountSection({
  children,
  description,
  id,
  title,
  variant = "default",
}: StorefrontAccountSectionProps) {
  return (
    <section
      aria-labelledby={`${id}-title`}
      className={cn(
        "storefront-account-section",
        variant === "danger" && "storefront-account-section--danger",
      )}
      id={id}
    >
      <header className="storefront-account-section__header">
        <h2 className="storefront-account-section__title" id={`${id}-title`}>
          {title}
        </h2>
        {description ? (
          <p className="storefront-account-section__lead">{description}</p>
        ) : null}
      </header>
      <div className="storefront-account-section__body">{children}</div>
    </section>
  );
}
