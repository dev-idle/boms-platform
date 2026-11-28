import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { PolicyDocument } from "./policy-document";
import { PolicySection } from "./policy-section";

/** The terms every account and every order is accepted under. */
export function TermsOfSale() {
  return (
    <PolicyDocument
      current="terms"
      lead={`The agreement between you and ${BRAND.name} when you open an account and when you order for pickup.`}
      title={PAGE_TITLES.terms}
    >
      <PolicySection id="terms-about" title="1. About these terms">
        <p>
          {BRAND.name} is a bakery in {BRAND.addressLine}. These terms apply when you create an
          account and to every order you place with us, online, at the counter or on the phone. You
          accept them when you sign up and again with each order you place online; we record which
          version you accepted and when.
        </p>
        <p>
          The <Link href={ROUTE.privacy}>{PAGE_TITLES.privacy}</Link> explains what we do with your
          personal data, and the <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link>{" "}
          explains cancellations, refunds and missed pickups. Both are part of these terms.
        </p>
      </PolicySection>

      <PolicySection id="terms-account" title="2. Your account">
        <ul>
          <li>Give us an email address you use and keep your details accurate.</li>
          <li>
            Confirm your email address with the link we send you, which works for 48 hours; you can
            ask for a new one on your account page. You need a confirmed address to order online,
            because every update about your orders goes there.
          </li>
          <li>Keep your password to yourself; you are responsible for orders placed with your account.</li>
          <li>
            You can update your details at any time, and delete your account from your account page,
            with your password, once none of your orders is waiting to be paid, made or collected.
          </li>
          <li>
            We may close an account used against these terms once none of its orders is still open;
            it is then signed out everywhere.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-ordering" title="3. Ordering and pickup">
        <ul>
          <li>
            Every order is collected at the bakery. We do not deliver, and we keep no delivery address.
          </li>
          <li>
            You choose a pickup day and time slot at checkout, within our opening hours and as many
            days ahead as checkout offers. We take no pickups on the days we are closed. Each slot
            takes a limited number of orders, and you can place at most three orders online for the
            same pickup day.
          </li>
          <li>
            Ready-made items collected on the day you order are an <strong>instant</strong> order and
            are packed shortly before pickup. Anything we bake to order makes the whole order a{" "}
            <strong>pre-order</strong>, which needs the notice shown at checkout.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-confirmed" title="4. When your order is confirmed">
        <ul>
          <li>
            An order you place online gets its order code at checkout, and we hold its pickup slot
            while you pay. Our agreement is made when PayPal confirms your payment, and we email you
            that we received your order. An order not paid within the time checkout gives you
            expires, and nothing is charged.
          </li>
          <li>A custom cake with a reference photo is confirmed only once we accept it (section 5).</li>
          <li>
            An order we take for you at the counter or on the phone is confirmed when we take it.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-custom" title="5. Custom cakes">
        <ul>
          <li>
            A custom cake is ordered online and made in the size, flavor, filling and decoration you
            choose, with the message you ask for. Without a reference photo it is confirmed when you
            pay.
          </li>
          <li>
            A cake you order with a reference photo is checked before we start: we accept or decline it
            within 24 hours of your payment, and one we have not accepted by then is cancelled. A cake
            we decline or cancel is refunded in full.
          </li>
          <li>
            A reference photo you upload must be one you took or may share. By uploading it you allow us
            to use it to make your cake; we do not publish it.
          </li>
          <li>We follow the photo as a guide: a cake made by hand will not match it exactly.</li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-prices" title="6. Prices and payment">
        <ul>
          <li>
            Prices are in US dollars and are the ones shown in your cart when you order. The total at
            checkout is what you pay: we add no fee to it.
          </li>
          <li>
            An order you place online is paid in full through PayPal, on PayPal&apos;s own page, when
            you place it. We never see or store your card details.
          </li>
          <li>
            An order we take for you at the counter or on the phone is paid in cash when you collect
            it.
          </li>
          <li>
            A discount code applies only to an order placed online and only under its own conditions,
            such as its dates, a minimum order and how many times it can be used; one code per order.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-our-changes" title="7. Changes we may make">
        <ul>
          <li>
            If we close on the day of your pickup, or change our opening hours so that your pickup time
            falls outside them, we cancel your order, refund what you paid in full and email you the
            reason.
          </li>
          <li>
            If we cannot make an order we accepted, we cancel it, tell you why and refund you in full.
          </li>
          <li>
            Our staff can move your pickup to another time we are open, when you ask them or when we
            cannot keep the one you booked. Your order page always shows the current pickup time.
          </li>
        </ul>
        <p>
          You can cancel or move an order yourself until we start making it; the{" "}
          <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link> explains how.
        </p>
      </PolicySection>

      <PolicySection id="terms-collection" title="8. Collecting your order">
        <p>
          Come in during your slot; we keep a ready order until closing time on its pickup day. For
          an order you placed online, give the 4-digit pickup code from your order page or our email
          at the counter: we hand the order over for its code. If you cannot show it, a manager can
          release the order once they are satisfied it is yours, and we keep a record of why. An
          order we took at the counter or on the phone has no code: it is handed over when you pay
          for it.
        </p>
        <p>
          An order that is ready but not collected by closing time on its pickup day is a missed
          pickup; the{" "}
          <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link> explains what happens
          then.
        </p>
      </PolicySection>

      <PolicySection id="terms-reviews" title="9. Reviews">
        <ul>
          <li>
            Once you pick an order up, you can review each product on it once: a rating of 1 to 5
            stars and, if you like, a comment.
          </li>
          <li>
            A review appears on the product&apos;s page as soon as you post it, signed with the first name
            and last initial on your profile, if you gave a name. We keep honest reviews, good or bad, and
            hide one that is abusive, shares someone&apos;s personal details or is not about the product.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-allergens" title="10. Ingredients and allergens">
        <p>
          Our kitchen handles gluten, milk, eggs, nuts and soy. Product pages describe each item
          but do not list every ingredient, and we cannot guarantee any item is free of traces. Ask
          us before ordering if you have an allergy.
        </p>
      </PolicySection>

      <PolicySection id="terms-liability" title="11. Our responsibility">
        <p>
          We make each order with care, as it is described on its product page and at checkout. If
          something is wrong with it, the{" "}
          <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link> explains what to do.
          Where we are responsible for a loss, our liability for an order is limited to the price
          you paid for it, except where the law does not allow such a limit. Nothing in these terms
          takes away the rights you have as a consumer under the law.
        </p>
      </PolicySection>

      <PolicySection id="terms-changes" title="12. Changes to these terms">
        <p>
          When we change these terms, the version and date at the top of this page change too, and
          you accept the new version with your next order. An order you already placed stays under
          the version you accepted for it.
        </p>
      </PolicySection>

      <PolicySection id="terms-contact" title="13. Law and contact">
        <p>
          These terms are governed by the law of Vietnam, where we trade. Pickup times on our site
          and in our emails are local time in Vietnam.
        </p>
        <p>
          Write to <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a> or call{" "}
          {BRAND.contactPhone}.
        </p>
      </PolicySection>
    </PolicyDocument>
  );
}
