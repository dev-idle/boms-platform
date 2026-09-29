import {
  BakerTicketDetail,
  bakerProductionDetailBreadcrumbItems,
} from "@/features/baker";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.ticket);

type BakerTicketPageProps = {
  params: Promise<{ id: string }>;
};

export default async function BakerTicketPage({ params }: BakerTicketPageProps) {
  const { id } = await params;

  return (
    <DashboardFormPage
      breadcrumbItems={bakerProductionDetailBreadcrumbItems()}
      description="Start the ticket, then mark it ready when it is done."
      title={PAGE_TITLES.ticket}
    >
      <BakerTicketDetail ticketId={id} />
    </DashboardFormPage>
  );
}
