"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { Button } from "@/components/ui/button";
import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardProfileFormSkeleton } from "@/components/ui/dashboard-profile-form-skeleton";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { useStoreSettings } from "../hooks";
import { ClosedDaysSection } from "./closed-days-section";
import { StoreSettingsForm } from "./store-settings-form";

function PickupRulesBody() {
  const settingsQuery = useStoreSettings();

  if (settingsQuery.isPending) {
    return <DashboardProfileFormSkeleton fields={7} label="Loading pickup rules" />;
  }
  if (settingsQuery.isError) {
    return (
      <div className="dashboard-profile-section-stack">
        <p className="text-error">Failed to load pickup rules.</p>
        <div className="dashboard-profile-form-actions">
          <Button type="button" variant="outline" onClick={() => void settingsQuery.refetch()}>
            Try again
          </Button>
        </div>
      </div>
    );
  }
  return <StoreSettingsForm settings={settingsQuery.data} />;
}

/** Admin settings: the pickup rules checkout enforces, changed without a deploy. */
export function AdminStoreSettings() {
  return (
    <DashboardFormPage
      description="Pickup hours, notice and closed days. Checkout follows every change at once."
      eyebrow={DASHBOARD_PAGE_EYEBROW.store}
      title={PAGE_TITLES.settings}
    >
      <DashboardProfileSection
        description="When customers can collect orders, how many each pickup slot takes, and how far ahead they can book."
        id="store-pickup-rules"
        title="Pickup rules"
        variant="plain"
      >
        <PickupRulesBody />
      </DashboardProfileSection>

      <DashboardProfileSection
        description="Days the bakery takes no pickups, such as holidays."
        id="store-closed-days"
        title="Closed days"
        variant="plain"
      >
        <ClosedDaysSection />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
