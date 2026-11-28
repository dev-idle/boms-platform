import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { ROUTE } from "@/constants/routes";
import { formatOrderChannelLabel, formatOrderTypeLabel } from "@/lib/schemas/order";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { formatVietnamPhone } from "@/lib/validation/phone";

import type { StaffOrder } from "../schemas";

const FACTS_TITLE_ID = "staff-order-facts-title";

type StaffOrderFactsProps = {
  order: StaffOrder;
};

/**
 * What an order is, as the detail page's sticky rail (archetype C): when it is
 * picked up and who it is for stay on screen while staff work the order.
 */
export function StaffOrderFacts({ order }: StaffOrderFactsProps) {
  const phone = formatVietnamPhone(order.customer.phone);

  return (
    <aside aria-labelledby={FACTS_TITLE_ID} className="dashboard-detail-rail">
      <h2 className="dashboard-detail-rail__title" id={FACTS_TITLE_ID}>
        Details
      </h2>
      <dl className="dashboard-detail-rail__facts">
        {order.pickup_at ? (
          <div>
            <dt>Pickup</dt>
            <dd className="dashboard-detail-rail__value--mono">{formatPickupDateTime(order.pickup_at)}</dd>
          </div>
        ) : null}
        <div>
          <dt>Placed</dt>
          <dd className="dashboard-detail-rail__value--mono">{formatDateTime(order.created_at)}</dd>
        </div>
        <div>
          <dt>Order type</dt>
          <dd>
            {formatOrderTypeLabel(order.order_type)}
            {order.channel === "online" ? null : ` · ${formatOrderChannelLabel(order.channel)}`}
          </dd>
        </div>
        <div className="dashboard-detail-rail__group">
          <dt>Customer</dt>
          {order.customer.display_name ? <dd>{order.customer.display_name}</dd> : null}
          {order.customer.email ? <dd>{order.customer.email}</dd> : null}
          {phone ? <dd className="dashboard-detail-rail__value--mono">{phone}</dd> : null}
          {/* A guest has no account to write to. */}
          {order.customer.email ? (
            <dd>
              <DashboardTableActionLink
                href={ROUTE.staff.chatThread(order.id)}
                label={`Message customer about ${order.code}`}
                showArrow
                text="Message customer"
              />
            </dd>
          ) : null}
        </div>
        {order.discount_code_snapshot ? (
          <div className="dashboard-detail-rail__group">
            <dt>Discount code</dt>
            <dd>
              <span className="text-order-code">{order.discount_code_snapshot}</span>
            </dd>
          </div>
        ) : null}
      </dl>
    </aside>
  );
}
