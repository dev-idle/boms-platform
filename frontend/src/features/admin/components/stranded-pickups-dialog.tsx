"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { InlineLoadingState } from "@/components/ui/loading-state";
import { formatOrderStatusLabel, orderStatusToPillVariant, StatusPill } from "@/components/ui/status-pill";
import { applyApiFormFieldErrors } from "@/lib/validation";
import { formatPickupDateTime } from "@/lib/validation/pickup";

import { useCancelStrandedPickups, useStrandedPickups } from "../hooks";
import { cancelStrandedFormSchema, type CancelStrandedFormInput, type StrandingChange } from "../schemas";

const FORM_ID = "admin-cancel-stranded";
const FIELDS = ["reason"] as const;

type StrandedPickupsDialogProps = {
  change: StrandingChange;
  /** The orders are cancelled: the change can be saved now. */
  onCleared: () => void;
  onClose: () => void;
};

/**
 * The orders a closed day or new opening hours would leave without a pickup,
 * for the admin to cancel — each refunded in full and its customer emailed the
 * reason — before the change is saved.
 */
export function StrandedPickupsDialog({ change, onCleared, onClose }: StrandedPickupsDialogProps) {
  const strandedQuery = useStrandedPickups(change);
  const cancel = useCancelStrandedPickups(change);
  const form = useForm<CancelStrandedFormInput>({
    resolver: zodResolver(cancelStrandedFormSchema),
    defaultValues: { reason: "" },
  });
  const stranded = strandedQuery.data;
  const what = "closed_date" in change ? "that day" : "the hours this change closes";
  const blocked = strandedQuery.isPending
    ? "Loading the orders"
    : strandedQuery.isError
      ? "The orders could not be loaded"
      : stranded?.total === 0
        ? "Nothing to cancel"
        : undefined;

  return (
    <AppDialog
      description={`Orders are booked for ${what}. Cancel them to go on: each customer reads your reason by email, and a paid order is refunded in full.`}
      footer={
        <AppDialogFooterActions>
          <Button disabled={cancel.isPending} type="button" variant="outline" onClick={onClose}>
            Keep orders
          </Button>
          <Button
            aria-busy={cancel.isPending || undefined}
            disabled={cancel.isPending || blocked !== undefined}
            form={FORM_ID}
            title={blocked}
            type="submit"
            variant="destructive"
          >
            {cancel.isPending
              ? "Cancelling…"
              : stranded
                ? `Cancel ${stranded.total} ${stranded.total === 1 ? "order" : "orders"}`
                : "Cancel orders"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={cancel.isPending}
      open
      panelClassName="app-dialog-panel--confirm app-dialog-panel--danger"
      title="Cancel the orders booked then?"
      onClose={onClose}
    >
      {strandedQuery.isPending ? <InlineLoadingState /> : null}
      {strandedQuery.isError ? (
        <div className="dashboard-profile-form-actions">
          <p className="text-sm text-error" role="alert">
            The orders could not be loaded.
          </p>
          <Button type="button" variant="outline" onClick={() => void strandedQuery.refetch()}>
            Try again
          </Button>
        </div>
      ) : null}
      {stranded?.total === 0 ? (
        <p className="text-sm text-muted">No orders are booked then any more: the change can be saved now.</p>
      ) : null}
      {stranded && stranded.total > 0 ? (
        <ul className="dashboard-order-line-items">
          {stranded.orders.map((order) => (
            <li key={order.id} className="dashboard-order-line-item">
              <div className="grid gap-1">
                <p className="text-order-code">{order.code}</p>
                {order.customer ? <p className="text-muted">{order.customer}</p> : null}
                <StatusPill label={formatOrderStatusLabel(order.status)} variant={orderStatusToPillVariant(order.status)} />
              </div>
              <p className="text-muted">{formatPickupDateTime(order.pickup_at)}</p>
            </li>
          ))}
          {stranded.total > stranded.orders.length ? (
            <li className="dashboard-order-line-item text-muted">
              and {stranded.total - stranded.orders.length} more
            </li>
          ) : null}
        </ul>
      ) : null}
      <Form {...form}>
        <form
          className="dashboard-profile-form"
          id={FORM_ID}
          method="post"
          noValidate
          onSubmit={form.handleSubmit((values) =>
            cancel.mutate(values, {
              onSuccess: onCleared,
              onError: (error) => {
                void strandedQuery.refetch();
                applyApiFormFieldErrors(form, error, FIELDS, "Failed to cancel the orders");
              },
            }),
          )}
        >
          <FormField
            control={form.control}
            name="reason"
            render={({ field }) => (
              <FormItem>
                <FieldControl hint="Up to 200 characters." hintId="admin-stranded-reason-hint" label="Reason">
                  <Input {...field} autoComplete="off" maxLength={200} />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>
    </AppDialog>
  );
}
