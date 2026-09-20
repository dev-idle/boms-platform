import { connection } from "next/server";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { BakerProductionTable, formatBakerShiftEyebrow } from "@/features/baker";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.production);

export default async function BakerProductionPage() {
  await connection();

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Kitchen queue sorted by pickup time."
        eyebrow={formatBakerShiftEyebrow(new Date())}
        leadAside
        title={PAGE_TITLES.production}
      />
      <BakerProductionTable />
    </div>
  );
}
