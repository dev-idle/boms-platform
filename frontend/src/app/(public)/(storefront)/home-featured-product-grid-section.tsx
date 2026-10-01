import { HomeFeaturedProductGrid } from "@/features/catalog";
import { ProductSaveButtons } from "@/features/customer";
import { dalListCatalogProducts, HOME_FEATURED_PRODUCT_COUNT } from "@/lib/dal/catalog";

export async function HomeFeaturedProductGridSection() {
  const products = await dalListCatalogProducts(1, HOME_FEATURED_PRODUCT_COUNT).catch(() => []);

  return (
    <HomeFeaturedProductGrid
      products={products}
      renderProductActions={(product) => <ProductSaveButtons product={product} />}
    />
  );
}
