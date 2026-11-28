"use client";

import { DashboardFormPage } from "@/components/ui/dashboard-form-page";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import {
  adminUsersBreadcrumb,
  CreateOperationalUserForm,
} from "@/features/admin";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

export default function AdminUsersNewPage() {
  return (
    <DashboardFormPage
      breadcrumbItems={adminUsersBreadcrumb(PAGE_TITLES.newUser)}
      description="Create staff, baker, or manager accounts. Admin accounts are not created here."
      title={PAGE_TITLES.newUser}
    >
      <DashboardProfileSection
        id="admin-users-new"
        title="Account details"
        variant="plain"
      >
        <CreateOperationalUserForm />
      </DashboardProfileSection>
    </DashboardFormPage>
  );
}
