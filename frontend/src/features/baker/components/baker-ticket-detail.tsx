"use client";

import { useState } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { CustomizationSummary } from "@/components/ui/customization-summary";
import { DashboardLoadError } from "@/components/ui/dashboard-load-error";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { InlineLoadingState } from "@/components/ui/loading-state";
import {
  formatOrderStatusLabel,
  StatusPill,
  ticketStatusToPillVariant,
} from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { STATION_LABEL } from "@/lib/schemas/ticket";
import { formatPickupDateTime } from "@/lib/validation/pickup";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

import { useKitchenTicket, usePatchKitchenTicketStatus } from "../hooks";
import { bakerProductionDetailBreadcrumbItems } from "../lib/baker-breadcrumbs";
import { nextTicketAction } from "../schemas";

type BakerTicketDetailProps = {
  ticketId: string;
};

/** A kitchen ticket to start and finish, beside where the rest of its order stands. */
export function BakerTicketDetail({ ticketId }: BakerTicketDetailProps) {
  const isValidId = z.uuid().safeParse(ticketId).success;
  const ticketQuery = useKitchenTicket(ticketId);
  const patchStatus = usePatchKitchenTicketStatus(ticketId);
  const [confirming, setConfirming] = useState(false);

  if (!isValidId) {
    return <p className="text-sm text-muted">This link does not lead to a ticket.</p>;
  }

  if (ticketQuery.isPending) {
    return <InlineLoadingState />;
  }

  if (ticketQuery.isError) {
    return isApiError(ticketQuery.error) && ticketQuery.error.status === 404 ? (
      <DashboardLoadError message="Ticket not found." />
    ) : (
      <DashboardLoadError message="We could not load this ticket." onRetry={() => void ticketQuery.refetch()} />
    );
  }

  const ticket = ticketQuery.data;
  if (!ticket) {
    return <p className="text-sm text-muted">Ticket not found.</p>;
  }

  const action = nextTicketAction(ticket.status);

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button
            aria-busy={patchStatus.isPending}
            disabled={!action || patchStatus.isPending}
            title={action ? undefined : "The ticket is ready for pickup"}
            type="button"
            onClick={() => setConfirming(true)}
          >
            {patchStatus.isPending ? "Updating…" : (action?.label ?? "Mark ready")}
          </Button>
        }
        breadcrumbItems={bakerProductionDetailBreadcrumbItems(ticket.order_code)}
        description={
          <StatusPill
            label={formatOrderStatusLabel(ticket.status)}
            variant={ticketStatusToPillVariant(ticket.status)}
          />
        }
        title={PAGE_TITLES.ticket}
      />

      {/* Archetype C: what to make on the left, the ticket's facts in the right rail. */}
      <div className="dashboard-detail-grid">
        <div className="dashboard-detail-grid__main">
          <DashboardProfileSection id="baker-ticket-items" title="Items" variant="plain">
            <ul className="dashboard-order-line-items">
              {/* A product can arrive twice — on its own and inside a combo — so the
                  line position is the key. */}
              {ticket.items.map((item, index) => (
                <li key={index} className="dashboard-order-line-item">
                  <div className="grid gap-1">
                    <p className="font-medium text-ink">
                      {item.quantity}× {item.name}
                    </p>
                    {item.customization ? <CustomizationSummary customization={item.customization} /> : null}
                  </div>
                </li>
              ))}
            </ul>
          </DashboardProfileSection>

          <DashboardProfileSection
            description="The order is ready for pickup when every station's ticket is."
            id="baker-order-tickets"
            title="Whole order"
            variant="plain"
          >
            <ul className="dashboard-order-line-items">
              {ticket.order_tickets.map((orderTicket) => (
                <li key={orderTicket.station} className="dashboard-order-line-item">
                  <p className="font-medium text-ink">{STATION_LABEL[orderTicket.station]}</p>
                  <StatusPill
                    label={formatOrderStatusLabel(orderTicket.status)}
                    variant={ticketStatusToPillVariant(orderTicket.status)}
                  />
                </li>
              ))}
            </ul>
          </DashboardProfileSection>
        </div>

        <aside aria-labelledby="baker-ticket-facts-title" className="dashboard-detail-rail">
          <h2 className="dashboard-detail-rail__title" id="baker-ticket-facts-title">
            Details
          </h2>
          <dl className="dashboard-detail-rail__facts">
            {ticket.pickup_at ? (
              <div>
                <dt>Pickup</dt>
                <dd className="dashboard-detail-rail__value--mono">{formatPickupDateTime(ticket.pickup_at)}</dd>
              </div>
            ) : null}
            <div>
              <dt>Order</dt>
              <dd>
                <span className="text-order-code">{ticket.order_code}</span>
              </dd>
            </div>
            {ticket.customer.display_name ? (
              <div>
                <dt>Customer</dt>
                <dd>{ticket.customer.display_name}</dd>
              </div>
            ) : null}
          </dl>
        </aside>
      </div>

      {action ? (
        <ConfirmDialog
          confirmLabel={action.label}
          confirmVariant="warning"
          description={`${ticket.order_code}: ${action.consequence}`}
          isPending={patchStatus.isPending}
          open={confirming}
          title={action.title}
          onCancel={() => setConfirming(false)}
          onConfirm={() => patchStatus.mutate({ status: action.status }, { onSettled: () => setConfirming(false) })}
        />
      ) : null}
    </div>
  );
}
