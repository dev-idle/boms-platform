import type { CatalogCategory, CatalogProduct } from "@/lib/schemas/catalog";
import { HomeCategories } from "./home-categories";
import { HomeCta } from "./home-cta";
import { HomeFeaturedProducts } from "./home-featured-products";
import { HomeHero } from "./home-hero";
import { HomeUspStrip } from "./home-usp-strip";

type StorefrontHomeProps = {
  categories: CatalogCategory[];
  products: CatalogProduct[];
};

export function StorefrontHome({ categories, products }: StorefrontHomeProps) {
  return (
    <>
      <HomeHero />
      <HomeUspStrip />
      <HomeCategories categories={categories} />
      <HomeFeaturedProducts products={products} />
      <HomeCta />
    </>
  );
}
