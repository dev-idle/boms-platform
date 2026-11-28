import { StaffOrderDetail } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orderDetail);

type StaffOrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function StaffOrderDetailPage({
  params,
}: StaffOrderDetailPageProps) {
  const { id } = await params;

  return <StaffOrderDetail orderId={id} />;
}
