import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { customerAccountSectionHref } from "@/constants/customer-account-sections";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { PolicyDocument } from "./policy-document";
import { PolicySection } from "./policy-section";

/** What personal data the bakery keeps, why, who sees it, and your rights over it. */
export function PrivacyPolicy() {
  return (
    <PolicyDocument
      current="privacy"
      lead="What we keep about you, why, who sees it, how long we keep it, and how to get it or have it removed."
      title={PAGE_TITLES.privacy}
    >
      <PolicySection id="privacy-who" title="1. Who looks after your data">
        <p>
          {BRAND.name}, {BRAND.addressLine}, decides how your personal data is used. For anything
          in this policy, write to <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a>.
        </p>
      </PolicySection>

      <PolicySection id="privacy-what" title="2. What we keep">
        <ul>
          <li>
            <strong>Your account:</strong> your email address and your password, which we store
            only as a one-way hash that nobody can read back.
          </li>
          <li>
            <strong>Your profile:</strong> the name and phone number you choose to give us.
          </li>
          <li>
            <strong>Your cart:</strong> the items you added and have not ordered yet.
          </li>
          <li>
            <strong>Your favorites and wishlist:</strong> the products you keep on them, and when.
          </li>
          <li>
            <strong>Your orders:</strong> what you ordered, the prices and any discount code, your
            pickup time, each step the order went through, and the version of our policies you
            accepted with it.
          </li>
          <li>
            <strong>Orders taken at the counter or by phone:</strong> if you order there without an
            account, the name and phone number you give us, so we can hand the order over and reach
            you about it.
          </li>
          <li>
            <strong>Messages:</strong> what you and our counter staff write to each other about one of
            your orders.
          </li>
          <li>
            <strong>Custom cakes:</strong> the options and message you choose, and any reference photo
            you upload. Cloudinary stores the photo at an unlisted address that we never publish.
          </li>
          <li>
            <strong>Payments:</strong> PayPal handles an online payment. We receive its reference and
            the amount, never your card or bank details. An order taken at the counter or by phone is
            paid in cash when you collect it.
          </li>
          <li>
            <strong>Signing in:</strong> your sign-in sessions with the device and network address
            they came from, so we can keep your account safe.
          </li>
          <li>
            <strong>Changes to your account:</strong> what changed and when, and for the changes you
            make, the network and browser you made them from.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="privacy-why" title="3. Why we keep it">
        <ul>
          <li>To take, make and hand over your orders — the agreement you make with us.</li>
          <li>To keep accounts secure and stop abuse, such as repeated failed sign-ins.</li>
          <li>To keep the sales records the law requires of a business.</li>
        </ul>
        <p>We do not sell your data and we do not send you marketing.</p>
      </PolicySection>

      <PolicySection id="privacy-who-sees" title="4. Who sees it">
        <ul>
          <li>
            Counter staff see your name, email and phone, to arrange and hand over your pickup, and the
            messages about your orders, to answer them.
          </li>
          <li>
            The kitchen sees what to make, when, your display name, and how you asked a custom cake to
            look, reference photo included — never how to contact you.
          </li>
          <li>Managers see sales figures, not your contact details.</li>
          <li>
            Administrators see your account and contact details, to look after accounts and the
            changes made to them.
          </li>
        </ul>
        <p>
          Our database is hosted by Neon, reference photos are stored by Cloudinary, and payments
          are processed by PayPal. They handle your data only to provide those services to us.
        </p>
      </PolicySection>

      <PolicySection id="privacy-how-long" title="5. How long we keep it">
        <ul>
          <li>
            Your account, profile, cart, favorites, wishlist and the messages about your orders, until
            you delete your account. Deleting it erases your name, phone number and email, empties your
            cart and your lists, erases those messages and removes your personal details from the
            record of changes; you are signed out everywhere and the account cannot be restored.
          </li>
          <li>Your sign-in sessions, until you sign out or they expire.</li>
          <li>
            A reference photo, as long as the order it was made for; one you uploaded without ordering
            stays stored until you ask us to delete it.
          </li>
          <li>
            Your orders, as part of our sales records, for as long as accounting law requires — after
            you delete your account, without your name or contact details.
          </li>
          <li>
            The name and phone number given for an order taken without an account, with that order in
            our sales records.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="privacy-rights" title="6. Your rights">
        <ul>
          <li>
            <strong>See your data:</strong> download everything we keep about you from your{" "}
            <Link href={customerAccountSectionHref("data")}>account page</Link>.
          </li>
          <li>
            <strong>Correct it:</strong> change your name and phone number on the same page.
          </li>
          <li>
            <strong>Have it removed:</strong> delete your account there once no order of yours is
            still waiting to be made or collected; your personal details are erased at once.
          </li>
          <li>
            <strong>Withdraw consent or object:</strong> write to us and we stop any use that relies
            on it.
          </li>
          <li>
            <strong>Complain:</strong> tell us first, and you can also contact the data protection
            authority where you live.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="privacy-cookies" title="7. Cookies and your browser">
        <p>
          We use only the cookies that keep you signed in, and your browser remembers your email
          for the current tab. We use no analytics or advertising cookies.
        </p>
      </PolicySection>

      <PolicySection id="privacy-changes" title="8. Changes to this policy">
        <p>
          The version and date at the top of this page change whenever this policy does. It is part
          of our <Link href={ROUTE.terms}>{PAGE_TITLES.terms}</Link>, which you accept with your next
          order.
        </p>
      </PolicySection>
    </PolicyDocument>
  );
}
