import Link from "next/link";

import { BRAND } from "@/constants/brand";
import { ROUTE } from "@/constants/routes";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { PolicyDocument } from "./policy-document";
import { PolicySection } from "./policy-section";

/** Cancellations, changes, refunds and missed pickups. */
export function RefundPolicy() {
  return (
    <PolicyDocument
      current="refunds"
      lead="When you can cancel or move a pickup, when you are refunded, and what happens if an order is not collected."
      title={PAGE_TITLES.refundPolicy}
    >
      <PolicySection id="refunds-before-baking" title="1. Before we start making your order">
        <ul>
          <li>
            An order you have not paid for yet can be cancelled at no cost. If you do not pay within
            the time checkout gives you, the order expires, nothing is charged and its pickup slot is
            released.
          </li>
          <li>
            Once paid, you can cancel an order from its order page until we start making it, and you
            are refunded in full. A discount code you used can then be used again, under its own
            conditions.
          </li>
          <li>Until then you can also move the pickup to another open slot.</li>
          <li>While PayPal is still reviewing your payment, you can cancel only once PayPal decides.</li>
        </ul>
      </PolicySection>

      <PolicySection id="refunds-in-production" title="2. Once we are making it">
        <p>
          When the kitchen or the counter has started on your order, it can no longer be cancelled
          or moved from your account. If we have to cancel it — for example because an ingredient
          ran out — we tell you why and refund you in full.
        </p>
      </PolicySection>

      <PolicySection id="refunds-custom" title="3. Custom cakes">
        <p>
          Custom cakes are paid in full when you order. A cake with a reference photo is checked
          before we start: if we cannot make it, or have not accepted it within 24 hours of your
          payment, we cancel it and refund you in full.
        </p>
      </PolicySection>

      <PolicySection id="refunds-closed" title="4. Closed days and new opening hours">
        <p>
          If we close on the day of your pickup, or change our opening hours so that your pickup time
          falls outside them, we cancel your order, refund what you paid in full and email you the
          reason.
        </p>
      </PolicySection>

      <PolicySection id="refunds-missed" title="5. Missed pickups">
        <ul>
          <li>
            An order that is ready but not collected by closing time on its pickup day is recorded as
            a missed pickup, and we email you. It is not refunded, because it was made for you and
            cannot be sold again. If you come in later, we can still hand it over for its pickup code.
          </li>
          <li>
            An order we took at the counter or on the phone is paid only when you collect it, so
            nothing is charged for a missed one.
          </li>
          <li>
            An order we did not have ready by closing time is not a missed pickup: we move it to
            another time or cancel it and refund what you paid in full.
          </li>
        </ul>
      </PolicySection>

      <PolicySection id="refunds-how" title="6. How refunds are paid">
        <p>
          Refunds go back through PayPal to the account or card you paid with, for the amount you
          paid. We send the refund to PayPal shortly after the order is cancelled: the cancellation
          email states the amount, and your order page shows the date the refund was made. If
          PayPal turns a refund down, we keep trying and a manager is told. When the money arrives
          is up to PayPal and your bank.
        </p>
        <p>
          If a payment goes through after its order has expired or been cancelled, we refund it in
          full.
        </p>
      </PolicySection>

      <PolicySection id="refunds-problems" title="7. Something wrong with your order">
        <p>
          Check your order at the counter. If something is missing or not right, tell us there, or
          contact us the same day at{" "}
          <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a>. Our staff record the
          problem with your order for a manager to review, and any money we return for an order
          paid online goes back through PayPal to the account you paid with. These rules are part
          of our <Link href={ROUTE.terms}>{PAGE_TITLES.terms}</Link>.
        </p>
      </PolicySection>
    </PolicyDocument>
  );
}
