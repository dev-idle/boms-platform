"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { FieldControl } from "@/components/ui/field-control";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { ASSIGNABLE_OPERATIONAL_ROLES, roleDisplayLabel } from "@/constants/roles";
import { DashboardProfileSection } from "@/features/user";
import { isApiError } from "@/lib/errors";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { applyFormFieldErrors } from "@/lib/validation";

import { useCreateOperational } from "../hooks";
import { adminUsersNewBreadcrumbItems } from "../lib/admin-breadcrumbs";
import {
  CREATE_OPERATIONAL_INITIAL,
} from "../lib/create-operational-form-values";
import { createOperationalSchema, type CreateOperationalInput } from "../schemas";

import { TempPasswordModal } from "./temp-password-modal";

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
          toast.error("That employee code is already in use.");
          return;
        }
        if (error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, [
            "email",
            "role",
            "full_name",
            "phone",
            "employee_code",
          ]);
          return;
        }
        toast.error(error.message);
      },
    });
  }

  return (
    <>
      <div className="dashboard-page-stack">
        <DashboardPageHeader
          breadcrumbItems={adminUsersNewBreadcrumbItems()}
          description="Create staff, baker, or manager accounts. Platform admins are created via dev seed only."
          title={PAGE_TITLES.newUser}
        />

        <div className="dashboard-page-body">
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
                        <Input autoComplete="email" type="email" {...field} />
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
                        <Input autoComplete="name" {...field} />
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
                          type="tel"
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
                        <Input autoComplete="off" {...field} />
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
