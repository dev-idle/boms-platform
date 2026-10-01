"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { IntegerFieldInput } from "@/components/ui/integer-field-input";
import { Select } from "@/components/ui/select";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { usePatchStoreSettings } from "../hooks";
import {
  MAX_ADVANCE_DAYS,
  MAX_INSTANT_PREP_MINUTES,
  MAX_PAYMENT_HOLD_MINUTES,
  MAX_PREORDER_LEAD_MINUTES,
  MAX_SLOT_CAPACITY,
  MIN_ADVANCE_DAYS,
  MIN_PAYMENT_HOLD_MINUTES,
  SLOT_MINUTE_OPTIONS,
  storeSettingsFormSchema,
  type StoreSettings,
  type StoreSettingsFormInput,
  type StoreSettingsPatch,
} from "../schemas";

const STORE_SETTINGS_FIELDS = [
  "opens_at",
  "closes_at",
  "slot_minutes",
  "slot_capacity",
  "preorder_min_lead_minutes",
  "instant_prep_minutes",
  "max_advance_days",
  "payment_hold_minutes",
] as const;

type StoreSettingsFormProps = {
  settings: StoreSettings;
};

function formValues(settings: StoreSettings): StoreSettingsFormInput {
  return {
    opens_at: settings.opens_at,
    closes_at: settings.closes_at,
    preorder_min_lead_minutes: settings.preorder_min_lead_minutes,
    max_advance_days: settings.max_advance_days,
    slot_minutes: settings.slot_minutes,
    slot_capacity: settings.slot_capacity,
    instant_prep_minutes: settings.instant_prep_minutes,
    payment_hold_minutes: settings.payment_hold_minutes,
  };
}

/** Opening hours, slots, notice, booking window and time to pay — checkout enforces them at once. */
export function StoreSettingsForm({ settings }: StoreSettingsFormProps) {
  const patchSettings = usePatchStoreSettings();
  const form = useForm<StoreSettingsFormInput>({
    resolver: zodResolver(storeSettingsFormSchema),
    defaultValues: formValues(settings),
  });

  function onSubmit(values: StoreSettingsFormInput): void {
    // Only what this admin changed: a save must not undo what another admin
    // saved to the other fields in the meantime.
    const { dirtyFields } = form.formState;
    const changes: StoreSettingsPatch = Object.fromEntries(
      STORE_SETTINGS_FIELDS.filter((field) => dirtyFields[field]).map((field) => [
        field,
        values[field],
      ]),
    );
    if (Object.keys(changes).length === 0) {
      toast.info("Nothing has changed");
      return;
    }
    patchSettings.mutate(changes, {
      onSuccess: (saved) => form.reset(formValues(saved)),
      onError: (error) => {
        if (isApiError(error) && error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, STORE_SETTINGS_FIELDS);
          return;
        }
        toast.error(isApiError(error) ? error.message : "Failed to save pickup rules");
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
          name="opens_at"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint="The first pickup time customers can choose, in bakery time."
                hintId="store-opens-at-hint"
                label="Pickups open"
              >
                <Input step={60} type="time" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="closes_at"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint="Pickups stop at this time; the last slot starts on the last mark before it."
                hintId="store-closes-at-hint"
                label="Pickups close"
              >
                <Input step={60} type="time" {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="slot_minutes"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint="Pickups start on these marks from opening time."
                hintId="store-slot-minutes-hint"
                label="Slot length"
              >
                <Select
                  name={field.name}
                  ref={field.ref}
                  value={String(field.value)}
                  onBlur={field.onBlur}
                  onChange={(event) => field.onChange(Number(event.target.value))}
                >
                  {SLOT_MINUTE_OPTIONS.map((minutes) => (
                    <option key={minutes} value={minutes}>
                      {minutes} minutes
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
          name="slot_capacity"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`How many orders one slot takes before it shows as full (1–${MAX_SLOT_CAPACITY}).`}
                hintId="store-slot-capacity-hint"
                label="Orders per slot"
              >
                <IntegerFieldInput max={MAX_SLOT_CAPACITY} min={1} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="preorder_min_lead_minutes"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`How long before pickup a pre-order must be placed, in minutes (at most ${MAX_PREORDER_LEAD_MINUTES}).`}
                hintId="store-lead-hint"
                label="Pre-order notice"
              >
                <IntegerFieldInput max={MAX_PREORDER_LEAD_MINUTES} min={0} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="instant_prep_minutes"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`How long the counter needs to pack a same-day order of ready-made items, in minutes (at most ${MAX_INSTANT_PREP_MINUTES}).`}
                hintId="store-instant-prep-hint"
                label="Instant pickup notice"
              >
                <IntegerFieldInput max={MAX_INSTANT_PREP_MINUTES} min={0} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="max_advance_days"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`How many days ahead customers can book a pickup (${MIN_ADVANCE_DAYS}–${MAX_ADVANCE_DAYS}).`}
                hintId="store-advance-hint"
                label="Booking window"
              >
                <IntegerFieldInput max={MAX_ADVANCE_DAYS} min={MIN_ADVANCE_DAYS} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="payment_hold_minutes"
          render={({ field }) => (
            <FormItem>
              <FieldControl
                hint={`How long an order waits for its PayPal payment before it expires and frees its pickup time, in minutes (${MIN_PAYMENT_HOLD_MINUTES}–${MAX_PAYMENT_HOLD_MINUTES}).`}
                hintId="store-payment-hold-hint"
                label="Time to pay"
              >
                <IntegerFieldInput max={MAX_PAYMENT_HOLD_MINUTES} min={MIN_PAYMENT_HOLD_MINUTES} {...field} />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="dashboard-profile-form-actions">
          <DashboardFormSaveButton
            idleLabel="Save pickup rules"
            isPending={patchSettings.isPending}
            pendingLabel="Saving…"
          />
        </div>
      </form>
    </Form>
  );
}
