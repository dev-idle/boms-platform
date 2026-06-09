import type { Metadata } from "next";
import { connection } from "next/server";
import { Suspense } from "react";
import { z } from "zod";

import { ProductDetail } from "@/features/catalog";
import { dalGetCatalogProduct } from "@/lib/dal/catalog";

type ProductDetailPageProps = {
  params: Promise<{ id: string }>;
};

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
        product.description ?? `Order ${product.name} for pickup at BOMS Bakery.`,
    };
  } catch {
    return { title: "Product" };
  }
}

export default async function ProductDetailPage({
  params,
}: ProductDetailPageProps) {
  const { id } = await params;

  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-7xl px-4 py-10 text-sm text-muted sm:px-6 lg:px-8">
          Loading product…
        </div>
      }
    >
      <ProductDetail productId={id} />
    </Suspense>
  );
}
