import {
  BakerProductionOrderDetail,
  bakerProductionDetailBreadcrumbItems,
} from "@/features/baker";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.orderDetail);

type BakerProductionOrderDetailPageProps = {
  params: Promise<{ id: string }>;
};

export default async function BakerProductionOrderDetailPage({
  params,
}: BakerProductionOrderDetailPageProps) {
  const { id } = await params;

  return (
    <DashboardFormPage
      breadcrumbItems={bakerProductionDetailBreadcrumbItems()}
      description="Review line items and advance production status."
      title={PAGE_TITLES.orderDetail}
    >
      <BakerProductionOrderDetail orderId={id} />
    </DashboardFormPage>
  );
}
