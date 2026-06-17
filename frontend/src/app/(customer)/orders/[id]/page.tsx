import { OrderDetail } from "@/features/customer";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orderDetail);

type OrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function OrderDetailPage({ params }: OrderDetailPageProps) {
  const { id } = await params;
  return (
    <div className="storefront-customer-section">
      <StorefrontPageHeader title={PAGE_TITLES.orderDetail} />
      <OrderDetail orderId={id} />
    </div>
  );
}
