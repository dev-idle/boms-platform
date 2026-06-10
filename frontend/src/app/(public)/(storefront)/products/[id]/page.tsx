import type { Metadata } from "next";
import { connection } from "next/server";
import { notFound } from "next/navigation";
import { z } from "zod";

import { BRAND } from "@/constants/brand";
import { dalGetCatalogProduct } from "@/lib/dal/catalog";
import { isApiError } from "@/lib/errors";
import type { CatalogProduct } from "@/lib/schemas/catalog";

import { ProductDetailPage } from "./product-detail-page";

type ProductDetailPageProps = {
  params: Promise<{ id: string }>;
};

async function loadCatalogProduct(id: string): Promise<CatalogProduct> {
  try {
    return await dalGetCatalogProduct(id);
  } catch (error) {
    if (isApiError(error) && error.status === 404) {
      notFound();
    }
    throw error;
  }
}

export async function generateMetadata({
  params,
}: ProductDetailPageProps): Promise<Metadata> {
  await connection();
  const { id } = await params;
  if (!z.string().uuid().safeParse(id).success) {
    return { title: "Product" };
  }

  try {
    const product = await dalGetCatalogProduct(id);
    return {
      title: product.name,
      description:
        product.description ??
        `Order ${product.name} for pickup at ${BRAND.name}.`,
    };
  } catch {
    return { title: "Product" };
  }
}

export default async function ProductDetailRoute({
  params,
}: ProductDetailPageProps) {
  const { id } = await params;

  if (!z.string().uuid().safeParse(id).success) {
    notFound();
  }

  const product = await loadCatalogProduct(id);
  return <ProductDetailPage initialProduct={product} productId={id} />;
}
