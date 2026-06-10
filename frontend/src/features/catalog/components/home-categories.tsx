import Link from "next/link";

import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";

import type { CatalogCategory } from "@/lib/schemas/catalog";

type HomeCategoriesProps = {
  categories: CatalogCategory[];
};

export function HomeCategories({ categories }: HomeCategoriesProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <section
      aria-labelledby="home-categories-heading"
      className="bg-bg py-24 sm:py-28"
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="reveal flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2
              className="font-heading text-3xl font-medium tracking-tight text-ink sm:text-4xl"
              id="home-categories-heading"
            >
              Browse by category
            </h2>
            <p className="mt-3 text-sm text-ink-2">
              From morning viennoiserie to celebration cakes.
            </p>
          </div>
        </div>

        <ul className="reveal reveal-delay-1 mt-10 flex gap-3 overflow-x-auto pb-2 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {categories.map((category) => (
            <li key={category.id} className="shrink-0">
              <Link
                className="inline-flex min-h-11 items-center rounded-full border border-border bg-surface px-5 py-3 text-sm font-medium text-ink transition-[background-color,box-shadow,color] duration-standard ease-default hover:bg-blush hover:text-rose-500"
                href={buildCatalogBrowseHref({ category: category.id, page: 1 })}
              >
                {category.name}
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
