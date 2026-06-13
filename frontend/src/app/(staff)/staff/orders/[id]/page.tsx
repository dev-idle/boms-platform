import { StaffOrderDetail } from "@/features/staff";
import { staffOrderDetailBreadcrumbItems } from "@/features/staff/lib/staff-breadcrumbs";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orderDetail);

type StaffOrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function StaffOrderDetailPage({
  params,
}: StaffOrderDetailPageProps) {
  const { id } = await params;

  return (
    <DashboardFormPage
      breadcrumbItems={staffOrderDetailBreadcrumbItems()}
      description="Review line items and update order status."
      title={PAGE_TITLES.orderDetail}
    >
      <StaffOrderDetail orderId={id} />
    </DashboardFormPage>
  );
}
