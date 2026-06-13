import { OrderDetail } from "@/features/customer";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orderDetail);

type OrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function OrderDetailPage({ params }: OrderDetailPageProps) {
  const { id } = await params;
  return (
    <div>
      <h1 className="text-page-title">{PAGE_TITLES.orderDetail}</h1>
      <div className="mt-6">
        <OrderDetail orderId={id} />
      </div>
    </div>
  );
}
