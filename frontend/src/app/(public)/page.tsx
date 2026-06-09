import type { Metadata } from "next";
import { connection } from "next/server";

import { StorefrontHome } from "@/features/catalog";
import {
  dalListCatalogCategories,
  dalListCatalogProducts,
} from "@/lib/dal/catalog";

export const metadata: Metadata = {
  title: "Home",
  description:
    "Handcrafted pastries and celebration cakes — order online for pickup at BOMS Bakery.",
};

export default async function HomePage() {
  await connection();

  const [categories, products] = await Promise.all([
    dalListCatalogCategories(8).catch(() => []),
    dalListCatalogProducts(1, 8).catch(() => []),
  ]);

  return <StorefrontHome categories={categories} products={products} />;
}
