"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { FormEvent } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ReadonlyFieldGroup } from "@/components/ui/readonly-field-group";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useHandOverOrder } from "../hooks";
import { handoffFormSchema, type HandoffFormInput } from "../schemas";

const FORM_ID = "staff-handoff";

type StaffHandoffDialogProps = {
  orderId: string;
  orderCode: string;
  /** What an order staff took is paid with in cash; null for an online order, handed over with its pickup code. */
  cashDueCents: number | null;
  open: boolean;
  onClose: () => void;
};

/** Hands a ready order over: an online one once the customer gives its pickup code, one staff took once its cash is paid. */
export function StaffHandoffDialog({ orderId, orderCode, cashDueCents, open, onClose }: StaffHandoffDialogProps) {
  const handOver = useHandOverOrder(orderId);
  const form = useForm<HandoffFormInput>({
    resolver: zodResolver(handoffFormSchema),
    defaultValues: { pickup_code: "" },
  });
  const paidInCash = cashDueCents !== null;

  function close(): void {
    form.reset();
    onClose();
  }

  function handOverWith(proof: { pickup_code: string } | { cash_collected: true }): void {
    handOver.mutate(proof, {
      onSuccess: close,
      onError: (error) => {
        if (isApiError(error) && error.code === ApiErrorCode.PickupCodeInvalid) {
          form.setError("pickup_code", { message: "That code does not match this order" });
          return;
        }
        if (isApiError(error) && error.code === ApiErrorCode.PickupCodeLocked) {
          form.setError("pickup_code", { message: "Too many wrong codes for this order. Try again later" });
          return;
        }
        toast.error(isApiError(error) ? error.message : "Failed to hand the order over");
      },
    });
  }

  function onSubmit(event: FormEvent<HTMLFormElement>): void {
    if (!paidInCash) {
      void form.handleSubmit(({ pickup_code }) => handOverWith({ pickup_code }))(event);
      return;
    }
    event.preventDefault();
    handOverWith({ cash_collected: true });
  }

  return (
    <AppDialog
      description={
        paidInCash
          ? "Take the cash due from the customer, then hand the order over."
          : "Ask the customer for the 4-digit pickup code on their order page or in their email."
      }
      footer={
        <AppDialogFooterActions>
          <Button disabled={handOver.isPending} type="button" variant="outline" onClick={close}>
            Not yet
          </Button>
          <Button aria-busy={handOver.isPending || undefined} disabled={handOver.isPending} form={FORM_ID} type="submit">
            {handOver.isPending ? "Handing over…" : paidInCash ? "Cash taken, hand over" : "Hand over"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={handOver.isPending}
      open={open}
      panelClassName="app-dialog-panel--confirm"
      title="Hand over this order?"
      onClose={close}
    >
      <Form {...form}>
        <form id={FORM_ID} method="post" noValidate onSubmit={onSubmit}>
          <p className="text-order-code">{orderCode}</p>
          {paidInCash ? (
            <ReadonlyFieldGroup
              heading="Cash due"
              reason="Set by the order's items when staff took it."
              rows={[{ label: "Total", value: formatPriceCents(cashDueCents), mono: true }]}
            />
          ) : (
            <>
              <FormField
                control={form.control}
                name="pickup_code"
                render={({ field }) => (
                  <FormItem>
                    <FieldControl label="Pickup code">
                      <Input autoComplete="off" inputMode="numeric" maxLength={4} placeholder="0000" {...field} />
                    </FieldControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </>
          )}
        </form>
      </Form>
    </AppDialog>
  );
}
