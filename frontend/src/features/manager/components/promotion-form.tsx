"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { usePromotionAudience, useSendPromotion } from "../hooks";
import {
  PROMOTION_BODY_MAX_LENGTH,
  PROMOTION_SUBJECT_MAX_LENGTH,
  promotionFormSchema,
  type PromotionFormInput,
} from "../schemas";

type PromotionFormProps = {
  onSent: () => void;
};

const FIELDS = ["subject", "body"] as const;

function customers(count: number): string {
  return count === 1 ? "1 customer" : `${count} customers`;
}

/**
 * A promotion's subject and message, sent once the manager confirms how many
 * customers it goes to. Emails cannot be called back, so sending always asks.
 */
export function PromotionForm({ onSent }: PromotionFormProps) {
  const audience = usePromotionAudience();
  const send = useSendPromotion();
  const [confirming, setConfirming] = useState<PromotionFormInput | null>(null);
  const form = useForm<PromotionFormInput>({
    resolver: zodResolver(promotionFormSchema),
    defaultValues: { subject: "", body: "" },
  });
  const recipients = audience.data?.recipients;

  function onConfirm(values: PromotionFormInput): void {
    send.mutate(values, {
      onSuccess: () => {
        setConfirming(null);
        onSent();
      },
      onError: (error) => {
        setConfirming(null);
        if (isApiError(error) && error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, FIELDS);
          return;
        }
        toast.error(isApiError(error) ? error.message : "Failed to send the promotion");
      },
    });
  }

  return (
    <Form {...form}>
      <form className="dashboard-profile-form" method="post" noValidate onSubmit={form.handleSubmit(setConfirming)}>
        <FormField
          control={form.control}
          name="subject"
          render={({ field }) => (
            <FormItem>
              <FieldControl hint={`The email's subject, up to ${PROMOTION_SUBJECT_MAX_LENGTH} characters.`} label="Subject">
                <Input placeholder="Matcha week: 10% off every cake" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="body"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`Plain text, up to ${PROMOTION_BODY_MAX_LENGTH} characters; each line is a paragraph. A link to the shop and one to unsubscribe are added below it.`}
                label="Message"
              >
                <textarea className="field-chrome text-form-input" rows={8} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {audience.isError ? (
          <div className="promotion-form__audience">
            <p className="form-field-hint text-error" role="status">
              We could not count who it goes to.
            </p>
            <Button type="button" variant="outline" onClick={() => void audience.refetch()}>
              Try again
            </Button>
          </div>
        ) : (
          <p className="form-field-hint" role="status">
            {recipients === undefined
              ? "Counting who it goes to…"
              : `Goes to ${customers(recipients)} who agreed to promotions and confirmed their address.`}
          </p>
        )}
        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            blockedReason={
              audience.isError
                ? "We could not count who it goes to"
                : recipients === undefined
                  ? "Counting who it goes to"
                  : recipients === 0
                    ? "No customer with a confirmed address has agreed to promotions yet"
                    : undefined
            }
            idleLabel="Send promotion"
            isPending={send.isPending}
            pendingLabel="Sending…"
          />
        </div>
      </form>
      <ConfirmDialog
        confirmLabel="Send"
        confirmVariant="warning"
        description={`It goes to ${customers(recipients ?? 0)} now, and an email cannot be called back.`}
        isPending={send.isPending}
        onCancel={() => setConfirming(null)}
        onConfirm={() => {
          if (confirming) {
            onConfirm(confirming);
          }
        }}
        open={confirming !== null}
        title="Send this promotion?"
      />
    </Form>
  );
}
