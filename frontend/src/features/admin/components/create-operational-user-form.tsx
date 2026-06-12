"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useFormState } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { FieldControl } from "@/components/ui/field-control";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  ASSIGNABLE_OPERATIONAL_ROLES,
  USER_ROLE,
  roleDisplayLabel,
} from "@/constants/roles";
import { DashboardProfileSection } from "@/features/user";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import { adminUsersNewBreadcrumbItems } from "../lib/admin-breadcrumbs";
import { createOperationalSchema, type CreateOperationalInput } from "../schemas";

import { TempPasswordModal } from "./temp-password-modal";

export function CreateOperationalUserForm() {
  const createUser = useCreateOperational();

  const form = useForm<CreateOperationalInput>({
    resolver: zodResolver(createOperationalSchema),
    defaultValues: {
      email: "",
      role: USER_ROLE.staff,
      full_name: "",
      phone: null,
      employee_code: "",
    },
  });
  const { isDirty } = useFormState({ control: form.control });

  function onSubmit(values: CreateOperationalInput): void {
    createUser.mutate(values, {
      onError: (error) => {
        if (!isApiError(error)) {
          toast.error("Failed to create user");
          return;
        }
        if (error.isEmployeeCodeExists()) {
          toast.error("That employee code is already in use.");
          return;
        }
        if (error.hasValidationDetails()) {
          for (const item of mapValidationDetailsToFormErrors(error.details!)) {
            const field = item.field as keyof CreateOperationalInput;
            if (field in values) {
              form.setError(field, { message: item.message });
            }
          }
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <>
      <div className="dashboard-page-stack dashboard-account-profile-page">
        <DashboardPageHeader
          breadcrumbItems={adminUsersNewBreadcrumbItems()}
          description="Create staff, baker, or manager accounts. Platform admins are created via dev seed only."
          title="New operational user"
        />

        <div className="dashboard-profile-stack">
          <DashboardProfileSection id="admin-users-new" title="Account details">
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
                        <Input type="email" {...field} />
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
                        <Input {...field} />
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
                          {...field}
                          value={field.value ?? ""}
                          onChange={(e) => field.onChange(e.target.value || null)}
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
                      <FieldControl label="Employee code">
                        <Input {...field} />
                      </FieldControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="dashboard-profile-form-actions">
                  <Button
                    disabled={createUser.isPending || !isDirty}
                    type="submit"
                  >
                    {createUser.isPending ? "Creating…" : "Create user"}
                  </Button>
                </div>
              </form>
            </Form>
          </DashboardProfileSection>
        </div>
      </div>

      <TempPasswordModal
        data={createUser.tempPasswordData}
        onClose={createUser.clearTempPasswordData}
        open={createUser.tempPasswordData !== null}
      />
    </>
  );
}
