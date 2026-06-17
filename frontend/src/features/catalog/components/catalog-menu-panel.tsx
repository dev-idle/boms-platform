import type { ReactNode } from "react";

type CatalogMenuPanelProps = {
  children: ReactNode;
};

export function CatalogMenuPanel({ children }: CatalogMenuPanelProps) {
  return <aside className="catalog-menu">{children}</aside>;
}
