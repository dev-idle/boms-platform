"use client";

import { use } from "react";

import { ProductDetail } from "@/features/customer";

type PageProps = {
  params: Promise<{ id: string }>;
};

export default function CustomerProductDetailPage({ params }: PageProps) {
  const { id } = use(params);
  return <ProductDetail productId={id} />;
}
