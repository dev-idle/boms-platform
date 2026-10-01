import { formatOrderChannelLabel, type OrderChannel } from "@/lib/schemas/order";

import type { StaffOrderCustomer } from "../schemas";

type StaffCustomerCellProps = {
  customer: StaffOrderCustomer;
  channel: OrderChannel;
};

/** Who an order is for, in a list: the name over the email, or a guest's name over where staff took the order. */
export function StaffCustomerCell({ customer, channel }: StaffCustomerCellProps) {
  const secondary = customer.display_name ? (customer.email ?? formatOrderChannelLabel(channel)) : null;
  return (
    <div className="db-table-stacked-cell min-w-0">
      <span className="db-table-cell-primary truncate">{customer.display_name ?? customer.email}</span>
      {secondary ? <span className="truncate text-caption-dashboard text-muted">{secondary}</span> : null}
    </div>
  );
}
