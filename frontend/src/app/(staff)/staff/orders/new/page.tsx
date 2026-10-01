import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { StaffNewOrder, staffNewOrderBreadcrumbItems } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.newOrder);

export default function StaffNewOrderPage() {
  return (
    <DashboardFormPage
      breadcrumbItems={staffNewOrderBreadcrumbItems()}
      description="Take an order at the counter or on the phone. It is paid in cash when it is collected."
      title={PAGE_TITLES.newOrder}
    >
      <StaffNewOrder />
    </DashboardFormPage>
  );
}
