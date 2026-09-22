"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { ASSIGNABLE_OPERATIONAL_ROLES, roleDisplayLabel } from "@/constants/roles";
import { DashboardFormSaveButton } from "@/components/ui/dashboard-form-save-button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { FieldControl } from "@/components/ui/field-control";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { isApiError } from "@/lib/errors";
import { useRemountingFormSnapshot } from "@/lib/hooks/use-remounting-form-snapshot";
import { applyApiFormFieldErrors } from "@/lib/validation";

import { OperationalEmployeeCodeField } from "./operational-employee-code-field";
import { useUpdateRole } from "../hooks";
import {
  adminUserToRoleFormValues,
} from "../lib/map-user-form-values";
import {
  updateRoleSchema,
  type AdminUser,
  type UpdateRoleInput,
} from "../schemas";

const ROLE_FORM_FIELDS = ["role", "full_name", "phone"] as const;

type AdminUserDetailRoleFormBodyProps = {
  assignedEmployeeCode?: string | null;
  userId: string;
  initialValues: UpdateRoleInput;
  onSaved: (values: UpdateRoleInput) => void;
};

function AdminUserDetailRoleFormBody({
  assignedEmployeeCode,
  userId,
  initialValues,
  onSaved,
}: AdminUserDetailRoleFormBodyProps) {
  const updateRole = useUpdateRole();
  const [confirmRole, setConfirmRole] = useState(false);
  const [pendingRoleValues, setPendingRoleValues] = useState<UpdateRoleInput | null>(
    null,
  );

  const form = useForm<UpdateRoleInput>({
    resolver: zodResolver(updateRoleSchema),
    defaultValues: initialValues,
  });

  function requestRoleSubmit(values: UpdateRoleInput): void {
    setPendingRoleValues(values);
    if (values.role === initialValues.role) {
      // Name or phone only: nothing is created or replaced, so nothing to confirm.
      submitRoleValues(values);
      return;
    }
    setConfirmRole(true);
  }

  function confirmRoleSubmit(): void {
    if (!pendingRoleValues) {
      setConfirmRole(false);
      return;
    }
    submitRoleValues(pendingRoleValues);
  }

  function submitRoleValues(values: UpdateRoleInput): void {
    updateRole.mutate(
      {
        id: userId,
        input: {
          role: values.role,
          full_name: values.full_name?.trim() || undefined,
          // Normalized by the schema; "" clears the stored phone.
          phone: values.phone,
        },
      },
      {
        onSuccess: (updatedUser) => {
          onSaved(adminUserToRoleFormValues(updatedUser));
          setConfirmRole(false);
          setPendingRoleValues(null);
        },
        onError: (error) => {
          setConfirmRole(false);
          if (isApiError(error) && error.isCannotModifySelf()) {
            toast.error("You cannot change your own role.");
            return;
          }
          if (isApiError(error) && error.isInvalidRoleTransition()) {
            toast.error("This role change is not allowed.");
            return;
          }
          applyApiFormFieldErrors(form, error, ROLE_FORM_FIELDS, "Failed to save changes");
        },
      },
    );
  }

  return (
    <>
      <Form {...form}>
        <form
          className="dashboard-profile-form"
          noValidate
          onSubmit={form.handleSubmit(requestRoleSubmit)}
        >
            {/* Read-only identifier first, then the fields the admin edits. */}
            <OperationalEmployeeCodeField assignedCode={assignedEmployeeCode} />

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
                  <FieldControl label="Full name" optional>
                    <Input autoComplete="name" {...field} value={field.value ?? ""} />
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

            <div className="dashboard-profile-form-actions">
              <DashboardFormSaveButton
                idleLabel="Save changes"
                isPending={updateRole.isPending}
                pendingLabel="Saving…"
              />
            </div>
          </form>
        </Form>

      <ConfirmDialog
        confirmLabel="Apply role change"
        description="They will be signed out everywhere and must sign in again. Their profile and employee code stay as they are."
        isPending={updateRole.isPending}
        onCancel={() => {
          setConfirmRole(false);
          setPendingRoleValues(null);
        }}
        onConfirm={confirmRoleSubmit}
        open={confirmRole}
        title="Change user role?"
      />
    </>
  );
}

type AdminUserDetailRoleFormProps = {
  assignedEmployeeCode?: string | null;
  userId: string;
  initialSnapshot: UpdateRoleInput;
};

function AdminUserDetailRoleForm({
  assignedEmployeeCode,
  userId,
  initialSnapshot,
}: AdminUserDetailRoleFormProps) {
  const { commitSnapshot, formKey, snapshot } =
    useRemountingFormSnapshot(initialSnapshot);

  return (
    <AdminUserDetailRoleFormBody
      key={formKey}
      assignedEmployeeCode={assignedEmployeeCode}
      initialValues={snapshot}
      onSaved={commitSnapshot}
      userId={userId}
    />
  );
}

type AdminUserDetailRoleTabProps = {
  userId: string;
  user: AdminUser;
};

export function AdminUserDetailRoleTab({ userId, user }: AdminUserDetailRoleTabProps) {
  const initialSnapshot = adminUserToRoleFormValues(user);

  return (
    <AdminUserDetailRoleForm
      key={`${user.id}-${user.updated_at}`}
      assignedEmployeeCode={user.employee_code}
      initialSnapshot={initialSnapshot}
      userId={userId}
    />
  );
}
