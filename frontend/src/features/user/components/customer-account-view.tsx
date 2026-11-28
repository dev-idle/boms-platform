import { CUSTOMER_ACCOUNT_SECTION } from "@/constants/customer-account-sections";
import { StorefrontAccountSection } from "@/components/layouts/storefront-account-layout";
import { StorefrontAccountShell } from "@/components/layouts/storefront-account-shell";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ChangePasswordForm } from "./change-password-form";
import { CustomerAccountProfileForm } from "./customer-account-profile-form";
import { DeleteAccountCard } from "./delete-account-card";
import { DownloadMyDataCard } from "./download-my-data-card";
import { EmailVerificationNotice } from "./email-verification-notice";

/** Customer self-service — profile, password, your data, and delete on one page. */
export function CustomerAccountView() {
  return (
    <StorefrontAccountShell title={PAGE_TITLES.account}>
      <div className="storefront-account-column">
        <EmailVerificationNotice reason="Confirm your email to place orders and get updates about them." />

        <StorefrontAccountSection
          id={CUSTOMER_ACCOUNT_SECTION.profile}
          title="Profile"
        >
          <CustomerAccountProfileForm />
        </StorefrontAccountSection>

        <StorefrontAccountSection
          description="Changing your password signs you out on every device."
          id={CUSTOMER_ACCOUNT_SECTION.password}
          title="Password"
        >
          <ChangePasswordForm formClassName="storefront-account-form" />
        </StorefrontAccountSection>

        <StorefrontAccountSection
          description="Download a copy of everything we keep about you."
          id={CUSTOMER_ACCOUNT_SECTION.data}
          title="Your data"
        >
          <DownloadMyDataCard />
        </StorefrontAccountSection>

        <StorefrontAccountSection
          description="Erase your account and the personal details we keep."
          id={CUSTOMER_ACCOUNT_SECTION.delete}
          title="Delete account"
          variant="danger"
        >
          <DeleteAccountCard />
        </StorefrontAccountSection>
      </div>
    </StorefrontAccountShell>
  );
}
