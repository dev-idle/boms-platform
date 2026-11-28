"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { applyApiFormFieldErrors } from "@/lib/validation";

import { useHandOverWithoutCode } from "../hooks";
import { managerHandoverInputSchema, type ManagerHandoverInput } from "../schemas";

const FORM_ID = "manager-handover-form";
const FIELDS = ["reason"] as const;

type ManagerHandoverDialogProps = {
  order: { id: string; code: string };
  onClose: () => void;
};

/** Hands an online order over without its code, keeping how the manager checked who collected it. */
export function ManagerHandoverDialog({ order, onClose }: ManagerHandoverDialogProps) {
  const handOver = useHandOverWithoutCode(order);
  const form = useForm<ManagerHandoverInput>({
    resolver: zodResolver(managerHandoverInputSchema),
    defaultValues: { reason: "" },
  });

  return (
    <AppDialog
      description="The order is marked picked up. Your name and reason are kept in the audit log."
      footer={
        <AppDialogFooterActions>
          <Button disabled={handOver.isPending} type="button" variant="outline" onClick={onClose}>
            Not yet
          </Button>
          <Button
            aria-busy={handOver.isPending || undefined}
            disabled={handOver.isPending}
            form={FORM_ID}
            type="submit"
            variant="warning"
          >
            {handOver.isPending ? "Handing over…" : "Hand over"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={handOver.isPending}
      open
      panelClassName="app-dialog-panel--confirm"
      title="Hand over this order without the code?"
      onClose={onClose}
    >
      <p className="text-order-code">{order.code}</p>
      <Form {...form}>
        <form
          className="dashboard-profile-form"
          id={FORM_ID}
          method="post"
          noValidate
          onSubmit={form.handleSubmit((values) =>
            handOver.mutate(values, {
              onSuccess: onClose,
              onError: (error) => applyApiFormFieldErrors(form, error, FIELDS, "We could not hand the order over. Please try again."),
            }),
          )}
        >
          <FormField
            control={form.control}
            name="reason"
            render={({ field }) => (
              <FormItem>
                <FieldControl
                  hint="For example, the buyer's ID and the PayPal receipt matched the order. Up to 200 characters."
                  hintId="manager-handover-reason-hint"
                  label="How you checked"
                >
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
