import { HomeFeaturedProductGrid, HomeSectionError } from "@/features/catalog";
import { ProductCardPurchase, ProductSaveButtons } from "@/features/customer";
import { dalListCatalogProducts, HOME_FEATURED_PRODUCT_COUNT } from "@/lib/dal/catalog";

export async function HomeFeaturedProductGridSection() {
  const products = await dalListCatalogProducts(1, HOME_FEATURED_PRODUCT_COUNT).catch(() => null);

  if (products === null) {
    return <HomeSectionError title="We could not load the products" />;
  }

  return (
    <HomeFeaturedProductGrid
      products={products}
      renderProductPurchase={(product) => <ProductCardPurchase product={product} />}
      renderProductSaves={(product) => <ProductSaveButtons product={product} />}
    />
  );
}
