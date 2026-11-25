"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ReadonlyFieldGroup } from "@/components/ui/readonly-field-group";
import { formatOrderStatusLabel } from "@/components/ui/status-pill";
import { isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useOrderLookup } from "../hooks";
import { managerOrderLookupSchema, type ManagerOrderLookup } from "../schemas";
import { ManagerHandoverDialog } from "./manager-handover-dialog";

type FoundOrder = NonNullable<ReturnType<typeof useOrderLookup>["data"]>;

/** Why this order is not handed over here, or undefined when it may be. */
function handoverBlockedReason(order: FoundOrder): string | undefined {
  if (order.channel !== "online") {
    return "The counter hands an order it took over with its cash";
  }
  switch (order.status) {
    case "ready":
    case "no_show":
      return undefined;
    case "fulfilled":
      return "Already picked up";
    case "cancelled":
      return "Cancelled";
    default:
      return "Not ready yet";
  }
}

/**
 * An online order whose customer cannot give its pickup code: the manager finds
 * it by its order code, checks who is collecting it, and hands it over.
 */
export function ManagerHandover() {
  const [code, setCode] = useState<string | null>(null);
  const [handing, setHanding] = useState(false);
  const lookup = useOrderLookup(code);
  const form = useForm<ManagerOrderLookup>({
    resolver: zodResolver(managerOrderLookupSchema),
    defaultValues: { code: "" },
  });
  const order = lookup.data;
  const blocked = order ? handoverBlockedReason(order) : undefined;

  return (
    <DashboardProfileSection
      description="When a customer cannot give their pickup code, find the order, check who is collecting it, and hand it over yourself."
      id="manager-handover"
      title="Hand over without the code"
      variant="plain"
    >
      <div className="dashboard-profile-section-stack">
        <Form {...form}>
          <form
            className="dashboard-profile-form"
            method="post"
            noValidate
            onSubmit={form.handleSubmit((values) => setCode(values.code))}
          >
            <FormField
              control={form.control}
              name="code"
              render={({ field }) => (
                <FormItem>
                  <FieldControl label="Order code">
                    <Input autoComplete="off" placeholder="CH-261125-001" {...field} />
                  </FieldControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="dashboard-profile-form-actions">
              <Button aria-busy={lookup.isFetching || undefined} disabled={lookup.isFetching} type="submit" variant="outline">
                {lookup.isFetching ? "Finding…" : "Find order"}
              </Button>
            </div>
          </form>
        </Form>

        <div aria-live="polite" role="status">
          {order ? (
            <div className="dashboard-profile-section-stack">
              <ReadonlyFieldGroup
                heading="Order found"
                reason="Check who is collecting it against the order before you hand it over."
                rows={[
                  { label: "Order", value: order.code },
                  { label: "Status", value: formatOrderStatusLabel(order.status) },
                  { label: "Customer", value: order.customer.display_name ?? order.customer.email ?? "—" },
                  { label: "Pickup", value: order.pickup_at ? formatPickupDateTime(order.pickup_at) : "—" },
                  { label: "Total", value: formatPriceCents(order.total_cents) },
                ]}
              />
              <div className="dashboard-profile-form-actions">
                <Button
                  aria-disabled={blocked ? true : undefined}
                  disabled={blocked !== undefined}
                  tabIndex={blocked ? -1 : undefined}
                  title={blocked}
                  type="button"
                  onClick={() => setHanding(true)}
                >
                  Hand over without the code
                </Button>
              </div>
            </div>
          ) : lookup.isError ? (
            <p className="text-sm text-error">
              {isApiError(lookup.error) && lookup.error.status === 404
                ? "No order has that code."
                : "The order could not be looked up. Try again."}
            </p>
          ) : null}
        </div>
      </div>

      {handing && order ? <ManagerHandoverDialog order={order} onClose={() => setHanding(false)} /> : null}
    </DashboardProfileSection>
  );
}
