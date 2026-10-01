"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { formatPriceCents } from "@/lib/validation/catalog";

import { usePatchStaffOrderStatus } from "../hooks";
import { cancelOrderFormSchema, type CancelOrderFormInput } from "../schemas";

const FORM_ID = "staff-cancel-order";

type StaffCancelOrderDialogProps = {
  orderId: string;
  open: boolean;
  /** What goes back to the customer's PayPal account; null when the order was not paid. */
  refundCents: number | null;
  onClose: () => void;
};

/** Cancels an order with the reason its customer reads in the app and the email. */
export function StaffCancelOrderDialog({ orderId, open, refundCents, onClose }: StaffCancelOrderDialogProps) {
  const patchStatus = usePatchStaffOrderStatus(orderId);
  const form = useForm<CancelOrderFormInput>({
    resolver: zodResolver(cancelOrderFormSchema),
    defaultValues: { reason: "" },
  });

  function close(): void {
    form.reset();
    onClose();
  }

  return (
    <AppDialog
      description={
        refundCents === null
          ? "The customer reads your reason in the app and by email."
          : `The customer reads your reason in the app and by email, and ${formatPriceCents(refundCents)} goes back to their PayPal account.`
      }
      footer={
        <AppDialogFooterActions>
          <Button disabled={patchStatus.isPending} type="button" variant="outline" onClick={close}>
            Keep order
          </Button>
          <Button
            aria-busy={patchStatus.isPending || undefined}
            disabled={patchStatus.isPending}
            form={FORM_ID}
            type="submit"
            variant="destructive"
          >
            {patchStatus.isPending ? "Cancelling…" : "Cancel order"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={patchStatus.isPending}
      open={open}
      panelClassName="app-dialog-panel--confirm app-dialog-panel--danger"
      title="Cancel this order?"
      onClose={close}
    >
      <Form {...form}>
        <form
          id={FORM_ID}
          method="post"
          noValidate
          onSubmit={form.handleSubmit(({ reason }) =>
            patchStatus.mutate({ status: "cancelled", reason }, { onSuccess: close }),
          )}
        >
          <FormField
            control={form.control}
            name="reason"
            render={({ field }) => (
              <FormItem>
                <FieldControl hint="Up to 200 characters." hintId="staff-cancel-reason-hint" label="Reason">
                  <Input maxLength={200} placeholder="Out of matcha today" {...field} />
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
