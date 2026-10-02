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
          {BRAND.name} is a bakery at {BRAND.addressLine}. These terms apply when you create an
          account and to every order you place with us. You accept them when you sign up and again
          with each order; we record which version you accepted and when.
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
          <li>Keep your password to yourself; you are responsible for orders placed with your account.</li>
          <li>
            You can update your details at any time, and delete your account from your account page
            once no order of yours is still open.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-ordering" title="3. Ordering and pickup">
        <ul>
          <li>
            Every order is collected at the bakery. We do not deliver, and we keep no delivery address.
          </li>
          <li>
            You choose a pickup day and time slot at checkout. Each slot takes a limited number of
            orders, and one customer may book at most three orders for the same day.
          </li>
          <li>
            Ready-made items collected on the day you order are an <strong>instant</strong> order and
            are packed shortly before pickup. Anything we bake to order makes the whole order a{" "}
            <strong>pre-order</strong>, which needs the notice shown at checkout.
          </li>
          <li>
            Your order is placed once checkout shows its order code. We may decline an order we cannot
            make; you are then refunded in full.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-custom" title="4. Custom cakes">
        <ul>
          <li>
            A custom cake is made in the size, flavor and decoration you choose, with the message you
            ask for. We check every custom order before we start and may decline one we cannot make as
            asked; you are then refunded in full.
          </li>
          <li>
            A reference photo you upload must be one you took or may share. By uploading it you allow us
            to use it to make your cake; we do not publish it.
          </li>
          <li>We follow the photo as a guide: a cake made by hand will not match it exactly.</li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-prices" title="5. Prices and payment">
        <ul>
          <li>Prices are in US dollars and are the ones shown in your cart when you order.</li>
          <li>
            You pay the full amount through PayPal when you place the order. We never see or store
            your card details.
          </li>
          <li>
            A discount code applies only under its own conditions and dates; one code per order.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-collection" title="6. Collecting your order">
        <p>
          Come in during your slot and give the 4-digit pickup code from your order page or our
          email at the counter: we hand an order over only for its code. An order that is ready but
          not collected by closing time on its pickup day is a missed pickup; the{" "}
          <Link href={ROUTE.refundPolicy}>{PAGE_TITLES.refundPolicy}</Link> explains what happens
          then.
        </p>
      </PolicySection>

      <PolicySection id="terms-reviews" title="7. Reviews">
        <ul>
          <li>
            Once you pick an order up, you can review each product on it once: a rating of 1 to 5
            stars and, if you like, a comment.
          </li>
          <li>
            A manager reads every review before it appears on the product&apos;s page, without your name.
            We publish honest reviews, good or bad, and hide one that is abusive, shares someone&apos;s
            personal details or is not about the product.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="terms-allergens" title="8. Ingredients and allergens">
        <p>
          Our kitchen handles gluten, milk, eggs, nuts and soy. Product pages describe what goes
          into each item, but we cannot guarantee any item is free of traces. Ask us before ordering
          if you have an allergy.
        </p>
      </PolicySection>

      <PolicySection id="terms-changes" title="9. Changes to these terms">
        <p>
          When we change these terms, the version and date at the top of this page change too, and
          you accept the new version with your next order. An order you already placed stays under
          the version you accepted for it.
        </p>
      </PolicySection>

      <PolicySection id="terms-contact" title="10. Contact">
        <p>
          Write to <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a> or call{" "}
          {BRAND.contactPhone}.
        </p>
      </PolicySection>
    </PolicyDocument>
  );
}
