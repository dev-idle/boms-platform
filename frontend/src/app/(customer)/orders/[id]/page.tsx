import { OrderDetail } from "@/features/customer";

type OrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function OrderDetailPage({ params }: OrderDetailPageProps) {
  const { id } = await params;
  return (
    <div>
      <h1 className="text-page-title">
        Order details
      </h1>
      <div className="mt-6">
        <OrderDetail orderId={id} />
      </div>
    </div>
  );
}
