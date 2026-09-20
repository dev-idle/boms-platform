import { Suspense } from "react";

import { BRAND } from "@/constants/brand";
import {
  HomeCategories,
  HomeCategoryGridSkeleton,
  HomeCta,
  HomeFeaturedProductGridSkeleton,
  HomeFeaturedProducts,
  HomeHero,
  HomeUspStrip,
} from "@/features/catalog";
import { HOME_FEATURED_PRODUCT_COUNT, STOREFRONT_CATEGORY_PAGE_SIZE } from "@/lib/dal/catalog";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

import { HomeCategoryGridSection } from "./home-category-grid-section";
import { HomeFeaturedProductGridSection } from "./home-featured-product-grid-section";

export const instant = true;

export const metadata = pageTitle(
  PAGE_TITLES.home,
  `Order online for pickup at ${BRAND.name}. ${BRAND.tagline}`,
);

export default function HomePage() {
  return (
    <>
      <HomeHero />
      <HomeUspStrip />
      <HomeCategories>
        <Suspense fallback={<HomeCategoryGridSkeleton count={STOREFRONT_CATEGORY_PAGE_SIZE} />}>
          <HomeCategoryGridSection />
        </Suspense>
      </HomeCategories>
      <HomeFeaturedProducts>
        <Suspense fallback={<HomeFeaturedProductGridSkeleton count={HOME_FEATURED_PRODUCT_COUNT} />}>
          <HomeFeaturedProductGridSection />
        </Suspense>
      </HomeFeaturedProducts>
      <HomeCta />
    </>
  );
}
