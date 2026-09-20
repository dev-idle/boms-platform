"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import type { RefObject } from "react";

import { ChevronDownIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { CATALOG_COMBOS_HEADING_ID } from "@/features/catalog";
import { cn } from "@/lib/utils";

type StorefrontHeaderNavProps = {
  menuOpen: boolean;
  menuPanelId: string;
  onToggleMenu: () => void;
  triggerRef: RefObject<HTMLButtonElement | null>;
};

/** Three fixed links in the toolbar; the category list lives behind the menu trigger. */
export function StorefrontHeaderNav({
  menuOpen,
  menuPanelId,
  onToggleMenu,
  triggerRef,
}: StorefrontHeaderNavProps) {
  const pathname = usePathname();
  const activeCategory = useSearchParams().get("category");
  const onBrowse = pathname === ROUTE.products;

  return (
    <nav aria-label="Shop categories" className="storefront-header-nav">
      <ul className="storefront-header-nav__list">
        <li>
          <Link
            aria-current={onBrowse && !activeCategory ? "page" : undefined}
            className="storefront-header-nav__link"
            href={ROUTE.products}
          >
            Shop all
          </Link>
        </li>
        <li>
          <button
            ref={triggerRef}
            aria-controls={menuOpen ? menuPanelId : undefined}
            aria-expanded={menuOpen}
            className={cn(
              "storefront-header-nav__link storefront-header-nav__trigger",
              (menuOpen || Boolean(activeCategory)) &&
                "storefront-header-nav__trigger--active",
            )}
            onClick={onToggleMenu}
            type="button"
          >
            Categories
            <ChevronDownIcon
              className={cn(
                "storefront-header-nav__chevron",
                menuOpen && "storefront-header-nav__chevron--open",
              )}
            />
          </button>
        </li>
        <li>
          <Link
            className="storefront-header-nav__link"
            href={`${ROUTE.products}#${CATALOG_COMBOS_HEADING_ID}`}
          >
            Combos
          </Link>
        </li>
      </ul>
    </nav>
  );
}
