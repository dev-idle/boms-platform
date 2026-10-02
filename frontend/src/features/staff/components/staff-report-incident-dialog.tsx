"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { AppDialog, AppDialogFooterActions } from "@/components/ui/app-dialog";
import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Select } from "@/components/ui/select";
import { INCIDENT_TYPE_LABEL } from "@/lib/schemas/incident";
import { applyApiFormFieldErrors } from "@/lib/validation";

import { useReportOrderIncident } from "../hooks";
import {
  INCIDENT_NOTE_MAX_LENGTH,
  reportedIncidentTypeSchema,
  reportIncidentInputSchema,
  type ReportIncidentInput,
} from "../schemas";

const FORM_ID = "staff-report-incident";
const FIELDS = ["type", "note"] as const;

type StaffReportIncidentDialogProps = {
  orderId: string;
  open: boolean;
  onClose: () => void;
};

/** Reports what went wrong with an order to the manager's incident log. */
export function StaffReportIncidentDialog({ orderId, open, onClose }: StaffReportIncidentDialogProps) {
  const report = useReportOrderIncident(orderId);
  const form = useForm<ReportIncidentInput>({
    resolver: zodResolver(reportIncidentInputSchema),
    defaultValues: { note: "" },
  });

  function close(): void {
    form.reset();
    onClose();
  }

  return (
    <AppDialog
      description="Managers read it in the incident log, with your name."
      footer={
        <AppDialogFooterActions>
          <Button disabled={report.isPending} type="button" variant="outline" onClick={close}>
            Cancel
          </Button>
          <Button
            aria-busy={report.isPending || undefined}
            disabled={report.isPending}
            form={FORM_ID}
            type="submit"
          >
            {report.isPending ? "Reporting…" : "Report issue"}
          </Button>
        </AppDialogFooterActions>
      }
      isPending={report.isPending}
      open={open}
      panelClassName="app-dialog-panel--confirm"
      title="Report an issue"
      onClose={close}
    >
      <Form {...form}>
        <form
          className="dashboard-profile-form"
          id={FORM_ID}
          method="post"
          noValidate
          onSubmit={form.handleSubmit((values) =>
            report.mutate(values, {
              onSuccess: close,
              onError: (error) => applyApiFormFieldErrors(form, error, FIELDS, "Failed to report the issue"),
            }),
          )}
        >
          <FormField
            control={form.control}
            name="type"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="What went wrong">
                  <Select {...field}>
                    <option value="">Choose one</option>
                    {reportedIncidentTypeSchema.options.map((type) => (
                      <option key={type} value={type}>
                        {INCIDENT_TYPE_LABEL[type]}
                      </option>
                    ))}
                  </Select>
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="note"
            render={({ field }) => (
              <FormItem>
                <FieldControl
                  hint={`What happened and what you did about it, up to ${INCIDENT_NOTE_MAX_LENGTH} characters.`}
                  hintId="staff-incident-note-hint"
                  label="Note"
                >
                  <textarea
                    className="field-chrome text-form-input"
                    placeholder="Two croissants short; packed them while the customer waited."
                    rows={3}
                    {...field}
                  />
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
