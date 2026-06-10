import type { Metadata } from "next";
import { connection } from "next/server";

import { BRAND } from "@/constants/brand";
import { StorefrontHome } from "@/features/catalog/components/storefront-home";
import {
  dalListCatalogCategories,
  dalListCatalogProducts,
} from "@/lib/dal/catalog";

export const metadata: Metadata = {
  title: "Home",
  description: `Order online for pickup at ${BRAND.name}. ${BRAND.tagline}`,
};

export default async function HomePage() {
  await connection();

  const [categories, products] = await Promise.all([
    dalListCatalogCategories(8).catch(() => []),
    dalListCatalogProducts(1, 8).catch(() => []),
  ]);

  return <StorefrontHome categories={categories} products={products} />;
}
