import type { ReactNode } from "react";

type PolicySectionProps = {
  children: ReactNode;
  id: string;
  title: string;
};

/** A numbered part of a policy, linkable by its id. */
export function PolicySection({ children, id, title }: PolicySectionProps) {
  return (
    <section className="storefront-policy__section" id={id}>
      <h2 className="storefront-policy__heading">{title}</h2>
      {children}
    </section>
  );
}
