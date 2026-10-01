import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { formatOrderStatusLabel, orderStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { ROUTE } from "@/constants/routes";
import { formatPriceCents } from "@/lib/validation/catalog";
import { formatVietnamPhone } from "@/lib/validation/phone";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import type { StaffOrder } from "../schemas";

type StaffChatOrderProps = {
  order: StaffOrder;
  customerEmail: string;
};

/** The order a conversation is about, beside it: a reply usually needs it. */
export function StaffChatOrder({ order, customerEmail }: StaffChatOrderProps) {
  return (
    <aside aria-label="Order and customer" className="staff-chat__rail">
      <p className="staff-chat__rail-label">Linked order</p>
      <p className="staff-chat__rail-code">{order.code}</p>
      <StatusPill label={formatOrderStatusLabel(order.status)} variant={orderStatusToPillVariant(order.status)} />
      <dl className="staff-chat__facts">
        <div>
          <dt>Pickup</dt>
          <dd>{order.pickup_at ? formatPickupDateTime(order.pickup_at) : "—"}</dd>
        </div>
        <div>
          <dt>Items</dt>
          <dd>
            {order.items.map((item) => (
              <span className="staff-chat__fact-line" key={item.id}>
                {item.quantity}× {item.name}
              </span>
            ))}
          </dd>
        </div>
        <div>
          <dt>Total</dt>
          <dd className="staff-chat__fact-money">{formatPriceCents(order.total_cents)}</dd>
        </div>
      </dl>
      <DashboardTableActionLink
        href={ROUTE.staff.orderDetail(order.id)}
        label={`Open order ${order.code}`}
        showArrow
        text="Open order"
      />
      <p className="staff-chat__rail-label staff-chat__rail-label--section">Customer</p>
      {order.customer.display_name ? <p className="staff-chat__rail-name">{order.customer.display_name}</p> : null}
      <p className="staff-chat__rail-contact">{customerEmail}</p>
      {order.customer.phone ? (
        <p className="staff-chat__rail-contact staff-chat__rail-contact--phone">{formatVietnamPhone(order.customer.phone)}</p>
      ) : null}
    </aside>
  );
}
