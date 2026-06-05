import { OrderDetail } from "@/features/customer";

type OrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function OrderDetailPage({ params }: OrderDetailPageProps) {
  const { id } = await params;
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
        Order details
      </h1>
      <div className="mt-6">
        <OrderDetail orderId={id} />
      </div>
    </div>
  );
}
