import { formatOrderChannelLabel, type OrderChannel } from "@/lib/schemas/order";

import type { StaffOrderCustomer } from "../schemas";

type StaffCustomerCellProps = {
  customer: StaffOrderCustomer;
  channel: OrderChannel;
};

/**
 * Who an order is for, in a list: the name over the email, or over where the
 * order was taken when there is no email — always two lines, so the rows of a
 * list stand the same height.
 */
export function StaffCustomerCell({ customer, channel }: StaffCustomerCellProps) {
  const secondary = (customer.display_name && customer.email) || formatOrderChannelLabel(channel);
  return (
    <div className="db-table-stacked-cell min-w-0">
      <span className="db-table-cell-primary truncate">{customer.display_name ?? customer.email}</span>
      <span className="truncate text-caption-dashboard text-muted">{secondary}</span>
    </div>
  );
}
