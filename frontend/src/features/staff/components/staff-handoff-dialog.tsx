"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ApiErrorCode, isApiError } from "@/lib/errors";

import { useHandOverOrder } from "../hooks";
import { handoffFormSchema, type HandoffFormInput } from "../schemas";

const FORM_ID = "staff-handoff";

type StaffHandoffDialogProps = {
  orderId: string;
  orderCode: string;
  open: boolean;
  onClose: () => void;
};

/** Hands a ready order over once the customer gives its pickup code. */
export function StaffHandoffDialog({ orderId, orderCode, open, onClose }: StaffHandoffDialogProps) {
  const handOver = useHandOverOrder(orderId);
  const form = useForm<HandoffFormInput>({
    resolver: zodResolver(handoffFormSchema),
    defaultValues: { pickup_code: "" },
  });

  function close(): void {
    form.reset();
    onClose();
  }

  function onSubmit({ pickup_code }: HandoffFormInput): void {
    handOver.mutate(pickup_code, {
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

  return (
    <AppDialog
      description="Ask the customer for the 4-digit pickup code on their order page or in their email."
      footer={
        <AppDialogFooterActions>
          <Button disabled={handOver.isPending} type="button" variant="outline" onClick={close}>
            Not yet
          </Button>
          <Button aria-busy={handOver.isPending || undefined} disabled={handOver.isPending} form={FORM_ID} type="submit">
            {handOver.isPending ? "Handing over…" : "Hand over"}
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
        <form id={FORM_ID} method="post" noValidate onSubmit={form.handleSubmit(onSubmit)}>
          <p className="text-order-code">{orderCode}</p>
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
        </form>
      </Form>
    </AppDialog>
  );
}
