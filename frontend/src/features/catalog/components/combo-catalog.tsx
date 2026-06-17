"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";

import { useCatalogCombos } from "../hooks";
import {
  CATALOG_COMBOS_HEADING_ID,
  CATALOG_COMBOS_SECTION_ID,
  resolveStorefrontCombosHeadingHash,
  scrollToStorefrontCombosSection,
} from "../lib/scroll-to-storefront-anchor";
import { CatalogPagination } from "./catalog-pagination";
import { ComboCard } from "./combo-card";

const COMBOS_PAGE_SIZE = 12;

type ComboCatalogProps = {
  renderPurchaseActions?: (comboId: string) => ReactNode;
};

export function ComboCatalog({ renderPurchaseActions }: ComboCatalogProps) {
  const [page, setPage] = useState(1);
  const filter = useMemo(
    () => ({ page, page_size: COMBOS_PAGE_SIZE }),
    [page],
  );
  const combosQuery = useCatalogCombos(filter);
  const combos = combosQuery.data?.combos ?? [];
  const pagination = combosQuery.data?.pagination;
  const sectionMounted =
    combosQuery.isPending || combosQuery.isError || combos.length > 0;

  useEffect(() => {
    if (!sectionMounted) {
      return;
    }

    const anchorId = resolveStorefrontCombosHeadingHash();
    if (!anchorId) {
      return;
    }

    requestAnimationFrame(() => scrollToStorefrontCombosSection());
  }, [sectionMounted, combosQuery.isPending]);

  if (combosQuery.isPending) {
    return null;
  }

  if (combosQuery.isError) {
    return (
      <ComboCatalogSection>
        <ComboCatalogHeading variant="minimal" />
        <ComboCatalogMain>
          <p className="text-caption text-error">Failed to load combo bundles.</p>
        </ComboCatalogMain>
      </ComboCatalogSection>
    );
  }

  if (combos.length === 0) {
    return null;
  }

  return (
    <ComboCatalogSection>
      <ComboCatalogHeading pagination={pagination} variant="full" />

      <ComboCatalogMain>
        <div className="catalog-combo-grid">
          {combos.map((combo) => (
            <ComboCard
              key={combo.id}
              combo={combo}
              renderPurchaseActions={renderPurchaseActions}
            />
          ))}
        </div>

        {pagination ? (
          <CatalogPagination
            className="catalog-pagination catalog-combos__pagination"
            onPageChange={setPage}
            page={page}
            totalPages={pagination.total_pages}
          />
        ) : null}
      </ComboCatalogMain>
    </ComboCatalogSection>
  );
}

type ComboCatalogSectionProps = {
  children: ReactNode;
};

function ComboCatalogSection({ children }: ComboCatalogSectionProps) {
  return (
    <section
      aria-labelledby={CATALOG_COMBOS_HEADING_ID}
      className="catalog-combos"
      id={CATALOG_COMBOS_SECTION_ID}
    >
      <div className="storefront-container catalog-combos__container">
        {children}
      </div>
    </section>
  );
}

function ComboCatalogMain({ children }: { children: ReactNode }) {
  return <div className="catalog-combos__main">{children}</div>;
}

type ComboCatalogHeadingProps = {
  pagination?: { total: number };
  variant: "full" | "minimal";
};

function ComboCatalogHeading({ pagination, variant }: ComboCatalogHeadingProps) {
  if (variant === "minimal") {
    return (
      <header
        className="catalog-combos__header"
        id={CATALOG_COMBOS_HEADING_ID}
      >
        <h2 className="catalog-combos__title">Combo bundles</h2>
      </header>
    );
  }

  return (
    <header className="catalog-combos__header" id={CATALOG_COMBOS_HEADING_ID}>
      <div className="catalog-combos__header-copy">
        <p className="catalog-combos__eyebrow">Limited time</p>
        <h2 className="catalog-combos__title">Combo bundles</h2>
        <p className="catalog-combos__lead">
          Curated sets with bundle pricing for pickup.
        </p>
      </div>
      {pagination ? (
        <p className="catalog-combos__count">
          {pagination.total} {pagination.total === 1 ? "bundle" : "bundles"}
        </p>
      ) : null}
    </header>
  );
}
