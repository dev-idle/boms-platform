import Link from "next/link";

import { buildCatalogBrowseHref } from "../lib/catalog-browse-params";

import type { CatalogCategory } from "../schemas";

type HomeCategoriesProps = {
  categories: CatalogCategory[];
};

export function HomeCategories({ categories }: HomeCategoriesProps) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby="home-categories-heading" className="py-16 sm:py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2
              className="font-heading text-2xl font-medium tracking-tight text-foreground sm:text-3xl"
              id="home-categories-heading"
            >
              Browse by category
            </h2>
            <p className="mt-2 text-sm text-muted">
              From morning viennoiserie to celebration cakes.
            </p>
          </div>
        </div>

        <ul className="mt-8 flex gap-3 overflow-x-auto pb-2 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {categories.map((category) => (
            <li key={category.id} className="shrink-0">
              <Link
                className="inline-flex rounded-md border border-border bg-surface px-5 py-3 text-sm font-medium text-foreground transition-colors duration-default ease-default hover:border-border-strong hover:bg-surface-alt"
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
