import type { CatalogCategory, CatalogProduct } from "../schemas";
import { HomeBrandStory } from "./home-brand-story";
import { HomeCategories } from "./home-categories";
import { HomeCta } from "./home-cta";
import { HomeFeaturedProducts } from "./home-featured-products";
import { HomeHero } from "./home-hero";

type StorefrontHomeProps = {
  categories: CatalogCategory[];
  products: CatalogProduct[];
};

export function StorefrontHome({ categories, products }: StorefrontHomeProps) {
  return (
    <>
      <HomeHero />
      <HomeCategories categories={categories} />
      <HomeFeaturedProducts products={products} />
      <HomeBrandStory />
      <HomeCta />
    </>
  );
}
