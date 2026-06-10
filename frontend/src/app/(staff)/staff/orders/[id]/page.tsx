import { StaffOrderDetail } from "@/features/staff";

type StaffOrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function StaffOrderDetailPage({
  params,
}: StaffOrderDetailPageProps) {
  const { id } = await params;
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-medium text-ink">
          Order detail
        </h1>
      </div>
      <StaffOrderDetail orderId={id} />
    </div>
  );
}
