"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { ASSIGNABLE_OPERATIONAL_ROLES, roleDisplayLabel } from "@/constants/roles";
import { FORM_FIELD_HINT } from "@/constants/dashboard-form-copy";
import { isApiError } from "@/lib/errors";
import { applyFormFieldErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import { CREATE_OPERATIONAL_INITIAL } from "../lib/create-operational-form-values";
import { createOperationalSchema, type CreateOperationalInput } from "../schemas";

import { TempPasswordModal } from "./temp-password-modal";

const CREATE_OPERATIONAL_FORM_FIELDS = [
  "email",
  "role",
  "full_name",
  "phone",
  "employee_code",
] as const;

export function CreateOperationalUserForm() {
  const createUser = useCreateOperational();

  const form = useForm<CreateOperationalInput>({
    resolver: zodResolver(createOperationalSchema),
    defaultValues: CREATE_OPERATIONAL_INITIAL,
  });

  function onSubmit(values: CreateOperationalInput): void {
    createUser.mutate(values, {
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to create user");
          return;
        }
        if (error.isEmployeeCodeExists()) {
          form.setError("employee_code", {
            message: "This employee code is already in use",
          });
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(
            form,
            error.details!,
            CREATE_OPERATIONAL_FORM_FIELDS,
          );
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <>
      <Form {...form}>
        <form
          className="dashboard-profile-form"
          noValidate
          onSubmit={form.handleSubmit(onSubmit)}
        >
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Email">
                  <Input
                    autoComplete="email"
                    placeholder="name@bakery.example"
                    type="email"
                    {...field}
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="role"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Role">
                  <Select {...field}>
                    {ASSIGNABLE_OPERATIONAL_ROLES.map((role) => (
                      <option key={role} value={role}>
                        {roleDisplayLabel(role)}
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
            name="full_name"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Full name">
                  <Input autoComplete="name" placeholder="Full name" {...field} />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="phone"
            render={({ field }) => (
              <FormItem>
                <FieldControl label="Phone" optional>
                  <Input
                    autoComplete="tel"
                    inputMode="tel"
                    placeholder="Phone number"
                    type="tel"
                    {...field}
                    value={field.value ?? ""}
                    onChange={(event) =>
                      field.onChange(event.target.value || null)
                    }
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="employee_code"
            render={({ field }) => (
              <FormItem>
                <FieldControl
                  hint={FORM_FIELD_HINT.operationalEmployeeCode}
                  hintId="create-operational-employee-code-hint"
                  label="Employee code"
                >
                  <Input
                    autoComplete="off"
                    placeholder="EMP-001"
                    {...field}
                  />
                </FieldControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="dashboard-profile-form-actions">
            <DashboardFormSaveButton
              idleLabel="Create user"
              isPending={createUser.isPending}
              pendingLabel="Creating…"
            />
          </div>
        </form>
      </Form>

      <TempPasswordModal
        data={createUser.tempPasswordData}
        onClose={createUser.clearTempPasswordData}
        open={createUser.tempPasswordData !== null}
      />
    </>
  );
}
