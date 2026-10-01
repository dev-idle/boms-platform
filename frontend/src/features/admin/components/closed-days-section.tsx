"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { ApiErrorCode, isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useClosedDates, useCreateClosedDate, useDeleteClosedDate } from "../hooks";
import { closedDayBounds, formatClosedDay } from "../lib/closed-days";
import {
  closedDateFormSchema,
  type ClosedDate,
  type ClosedDateFormInput,
} from "../schemas";

const CLOSED_DAY_FIELDS = ["date", "reason"] as const;

function AddClosedDayForm() {
  const createClosedDate = useCreateClosedDate();
  const bounds = closedDayBounds();
  const form = useForm<ClosedDateFormInput>({
    resolver: zodResolver(closedDateFormSchema),
    defaultValues: { date: "", reason: "" },
  });

  function onSubmit(values: ClosedDateFormInput): void {
    createClosedDate.mutate(values, {
      onSuccess: () => form.reset({ date: "", reason: "" }),
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to add the closed day");
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, CLOSED_DAY_FIELDS);
          return;
        }
        if (error.code === ApiErrorCode.ClosedDateExists) {
          form.setError("date", { message: "That day is already closed" });
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <Form {...form}>
      <form
        method="post"
        className="dashboard-profile-form"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="date"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint="A day in bakery time, from today to a year ahead."
                hintId="closed-day-date-hint"
                label="Day"
              >
                <Input max={bounds.max} min={bounds.min} type="date" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="reason"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint="Customers see this when they pick that day."
                hintId="closed-day-reason-hint"
                label="Reason"
              >
                <Input maxLength={200} placeholder="Lunar New Year" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel="Close this day"
            isPending={createClosedDate.isPending}
            pendingLabel="Closing…"
            variant="outline"
          />
        </div>
      </form>
    </Form>
  );
}

type ClosedDaysTableProps = {
  closedDates: ClosedDate[];
  initialLoading: boolean;
  isError: boolean;
  refetching: boolean;
};

function ClosedDaysTable({ closedDates, initialLoading, isError, refetching }: ClosedDaysTableProps) {
  const deleteClosedDate = useDeleteClosedDate();
  const [reopening, setReopening] = useState<ClosedDate | null>(null);

  return (
    <>
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--closed-days">
          <colgroup>
            <col className="db-table-col-datetime" />
            <col />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th>Day</th>
              <th>Reason</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={3}
              emptyMessage="No closed days are planned."
              entityLabel="closed days"
              initialLoading={initialLoading}
              isEmpty={closedDates.length === 0}
              isError={isError}
            />
            {closedDates.map((closed) => (
              <tr key={closed.id}>
                <td className="db-table-datetime">
                  <time dateTime={closed.date}>
                    <span className="db-table-datetime__date">{formatClosedDay(closed.date)}</span>
                  </time>
                </td>
                <td>{closed.reason}</td>
                <td className="db-table-detail">
                  <DashboardTableActionButton
                    label={`Reopen ${formatClosedDay(closed.date)}`}
                    text="Reopen"
                    tone="warning"
                    onClick={() => setReopening(closed)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </DashboardTableWrap>
      <ConfirmDialog
        confirmLabel="Reopen day"
        confirmVariant="warning"
        description={
          reopening
            ? `Customers will be able to book pickups on ${formatClosedDay(reopening.date)} again.`
            : ""
        }
        isPending={deleteClosedDate.isPending}
        open={reopening !== null}
        title="Reopen this day?"
        onCancel={() => setReopening(null)}
        onConfirm={() => {
          if (!reopening) {
            return;
          }
          deleteClosedDate.mutate(reopening.id, {
            onSettled: () => setReopening(null),
            onError: (error) => {
              toast.error(isApiError(error) ? error.message : "Failed to reopen the day");
            },
          });
        }}
      />
    </>
  );
}

/** The days the bakery takes no pickups: add one, or reopen one. */
export function ClosedDaysSection() {
  const closedDatesQuery = useClosedDates();

  return (
    <div className="dashboard-profile-section-stack">
      <AddClosedDayForm />
      <ClosedDaysTable
        closedDates={closedDatesQuery.data ?? []}
        initialLoading={closedDatesQuery.isPending}
        isError={closedDatesQuery.isError}
        refetching={closedDatesQuery.isFetching && !closedDatesQuery.isPending}
      />
    </div>
  );
}
