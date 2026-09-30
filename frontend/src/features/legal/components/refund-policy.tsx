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
            the time checkout gives you, the order expires and its pickup slot is released.
          </li>
          <li>
            Once paid, you can cancel an order until we start making it and you are refunded in full.
            A discount code you used can then be used again.
          </li>
          <li>Until then you can also move the pickup to another open slot.</li>
        </ul>
      </PolicySection>

      <PolicySection id="refunds-in-production" title="2. Once we are making it">
        <p>
          When the kitchen or the counter has started on your order, it can no longer be cancelled
          or moved from your account. If we have to cancel it — for example because an ingredient
          ran out — we tell you why and refund you in full.
        </p>
      </PolicySection>

      <PolicySection id="refunds-custom" title="3. Made-to-order cakes">
        <p>
          Made-to-order cakes are paid in full when you order. We review each request before
          baking; if we cannot make it, we decline it and refund you in full.
        </p>
      </PolicySection>

      <PolicySection id="refunds-missed" title="4. Missed pickups">
        <p>
          An order that is ready but not collected by closing time on its pickup day is recorded as
          a missed pickup. It is not refunded, because it was made for you and cannot be sold again.
        </p>
      </PolicySection>

      <PolicySection id="refunds-how" title="5. How refunds are paid">
        <p>
          Refunds go back through PayPal to the account or card you paid with, for the amount you
          paid. When the money arrives is up to PayPal and your bank.
        </p>
      </PolicySection>

      <PolicySection id="refunds-problems" title="6. Something wrong with your order">
        <p>
          Check your order at the counter. If something is missing or not right, tell us there, or
          contact us the same day at{" "}
          <a href={`mailto:${BRAND.contactEmail}`}>{BRAND.contactEmail}</a>. These rules are part of
          our <Link href={ROUTE.terms}>{PAGE_TITLES.terms}</Link>.
        </p>
      </PolicySection>
    </PolicyDocument>
  );
}
