"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PhoneInput } from "@/components/ui/phone-input";
import { Select } from "@/components/ui/select";
import { ASSIGNABLE_OPERATIONAL_ROLES, roleDisplayLabel } from "@/constants/roles";
import { isApiError } from "@/lib/errors";
import { applyApiFormFieldErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import { CREATE_OPERATIONAL_INITIAL } from "../lib/create-operational-form-values";
import { createOperationalSchema, type CreateOperationalInput } from "../schemas";

import { OperationalEmployeeCodeField } from "./operational-employee-code-field";
import { TempPasswordModal } from "./temp-password-modal";

const CREATE_OPERATIONAL_FORM_FIELDS = [
  "email",
  "role",
  "full_name",
  "phone",
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
        if (isApiError(error) && error.isEmailExists()) {
          form.setError("email", { message: "An account with this email already exists" });
          return;
        }
        applyApiFormFieldErrors(
          form,
          error,
          CREATE_OPERATIONAL_FORM_FIELDS,
          "Failed to create user",
        );
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
          {/* Read-only identifier first, then the fields the admin edits. */}
          <OperationalEmployeeCodeField />

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
                  {/* The admin is entering someone else's number: the
                      browser's own phone history would be the wrong one. */}
                  <PhoneInput autoComplete="off" {...field} />
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
