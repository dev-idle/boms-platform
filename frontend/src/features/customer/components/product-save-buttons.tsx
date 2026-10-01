"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

import { BookmarkIcon, HeartIcon } from "@/components/icons/storefront-icons";
import type { CatalogProduct } from "@/lib/schemas/catalog";
import { loginHrefPreservingNext } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { useSavedProducts, useToggleSaved } from "../hooks";
import { SAVED_LIST_NOUN, type SavedList } from "../schemas";

type ProductSaveButtonsProps = {
  product: CatalogProduct;
};

const LISTS: ReadonlyArray<{ list: SavedList; title: string; label: (name: string) => string; Icon: typeof HeartIcon }> = [
  { list: "favorite", title: "Favorite", label: (name) => `Favorite ${name}`, Icon: HeartIcon },
  { list: "wishlist", title: "Wishlist", label: (name) => `Save ${name} to your wishlist`, Icon: BookmarkIcon },
];

/**
 * The heart and the bookmark that keep a product on the customer's favorites
 * and wishlist, filled while it is on them. A guest is asked to sign in first,
 * and comes back to the page they were on.
 */
export function ProductSaveButtons({ product }: ProductSaveButtonsProps) {
  const signedIn = useAuthStore((state) => state.status === "authenticated");
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const savedQuery = useSavedProducts({ enabled: signedIn });
  const toggle = useToggleSaved();
  const search = searchParams.toString() ? `?${searchParams.toString()}` : "";

  return (
    <div className="saved-toggles">
      {LISTS.map(({ list, title, label, Icon }) => {
        if (!signedIn) {
          return (
            <Link
              aria-label={`Sign in to keep ${product.name} in your ${SAVED_LIST_NOUN[list]}`}
              className="saved-toggle"
              href={loginHrefPreservingNext(pathname, search)}
              key={list}
              title={`Sign in to use your ${SAVED_LIST_NOUN[list]}`}
            >
              <Icon />
            </Link>
          );
        }
        const onList = savedQuery.data?.some((item) => item.list === list && item.product.id === product.id) ?? false;
        return (
          <button
            aria-label={label(product.name)}
            aria-pressed={onList}
            className="saved-toggle"
            disabled={!savedQuery.isSuccess || toggle.isPending}
            key={list}
            title={savedQuery.isError ? "We could not load your lists" : title}
            type="button"
            onClick={() => toggle.mutate({ list, product, onList })}
          >
            <Icon />
          </button>
        );
      })}
    </div>
  );
}
