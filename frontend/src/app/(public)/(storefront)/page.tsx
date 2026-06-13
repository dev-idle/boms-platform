import { connection } from "next/server";

import { BRAND } from "@/constants/brand";
import { StorefrontHome } from "@/features/catalog/components/storefront-home";
import {
  dalListCatalogCategories,
  dalListCatalogProducts,
} from "@/lib/dal/catalog";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(
  PAGE_TITLES.home,
  `Order online for pickup at ${BRAND.name}. ${BRAND.tagline}`,
);

export default async function HomePage() {
  await connection();

  const [categories, products] = await Promise.all([
    dalListCatalogCategories(8).catch(() => []),
    dalListCatalogProducts(1, 8).catch(() => []),
  ]);

  return <StorefrontHome categories={categories} products={products} />;
}
