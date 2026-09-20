import { CUSTOMER_ACCOUNT_SECTION } from "@/constants/customer-account-sections";
import { StorefrontPageHeader } from "@/components/layouts/storefront-page-header";
import { StorefrontAccountSection } from "@/components/layouts/storefront-account-layout";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ChangePasswordForm } from "./change-password-form";
import { CustomerAccountSignedInAs } from "./customer-account-signed-in-as";
import { CustomerAccountProfileForm } from "./customer-account-profile-form";
import { DeleteAccountCard } from "./delete-account-card";

/** Customer self-service — profile, password, and delete on one page. */
export function CustomerAccountView() {
  return (
    <div className="storefront-customer-section storefront-customer-section--account">
      <StorefrontPageHeader
        eyebrow="Your account"
        lead="Manage your contact details, password, and account preferences in one place."
        title={PAGE_TITLES.account}
      />

      <CustomerAccountSignedInAs />

      <div>
        <StorefrontAccountSection
          id={CUSTOMER_ACCOUNT_SECTION.profile}
          title="Profile"
        >
          <CustomerAccountProfileForm />
        </StorefrontAccountSection>

        <StorefrontAccountSection
          description="Changing your password signs you out of all active sessions."
          id={CUSTOMER_ACCOUNT_SECTION.password}
          title="Password"
        >
          <ChangePasswordForm formClassName="storefront-account-form" />
        </StorefrontAccountSection>

        <StorefrontAccountSection
          description="Permanently remove your customer account and order access."
          id={CUSTOMER_ACCOUNT_SECTION.delete}
          title="Delete account"
          variant="danger"
        >
          <DeleteAccountCard />
        </StorefrontAccountSection>
      </div>
    </div>
  );
}
