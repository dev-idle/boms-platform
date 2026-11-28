import { BakerTicketDetail } from "@/features/baker";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.ticket);

type BakerTicketPageProps = {
  params: Promise<{ id: string }>;
};

export default async function BakerTicketPage({ params }: BakerTicketPageProps) {
  const { id } = await params;

  return <BakerTicketDetail ticketId={id} />;
}
