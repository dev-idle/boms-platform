"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { ASSIGNABLE_OPERATIONAL_ROLES, USER_ROLE } from "@/constants/roles";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { isApiError } from "@/lib/errors";

import { useUpdateRole } from "../hooks";
import { applyFormFieldErrors } from "../lib/apply-form-validation";
import { adminUserToRoleFormValues } from "../lib/map-user-form-values";
import {
  updateRoleSchema,
  type AdminUser,
  type UpdateRoleInput,
} from "../schemas";

type AdminUserDetailRoleFormProps = {
  userId: string;
  user: AdminUser;
};

function AdminUserDetailRoleForm({ userId, user }: AdminUserDetailRoleFormProps) {
  const updateRole = useUpdateRole();
  const [confirmRole, setConfirmRole] = useState(false);
  const [pendingRoleValues, setPendingRoleValues] = useState<UpdateRoleInput | null>(
    null,
  );

  const form = useForm<UpdateRoleInput>({
    resolver: zodResolver(updateRoleSchema),
    defaultValues: adminUserToRoleFormValues(user),
  });

  function requestRoleSubmit(values: UpdateRoleInput): void {
    setPendingRoleValues(values);
    setConfirmRole(true);
  }

  function confirmRoleSubmit(): void {
    if (!pendingRoleValues) {
      setConfirmRole(false);
      return;
    }

    updateRole.mutate(
      {
        id: userId,
        input: {
          role: pendingRoleValues.role,
          full_name: pendingRoleValues.full_name?.trim() || undefined,
          phone: pendingRoleValues.phone?.trim()
            ? pendingRoleValues.phone.trim()
            : null,
          employee_code: pendingRoleValues.employee_code?.trim()
            ? pendingRoleValues.employee_code.trim()
            : null,
          hire_date: pendingRoleValues.hire_date?.trim()
            ? pendingRoleValues.hire_date.trim()
            : null,
          shift: pendingRoleValues.shift?.trim()
            ? pendingRoleValues.shift.trim()
            : null,
        },
      },
      {
        onSuccess: () => {
          setConfirmRole(false);
          setPendingRoleValues(null);
        },
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Failed to update role");
            return;
          }
          if (error.hasValidationDetails() && pendingRoleValues) {
            applyFormFieldErrors(form, error.details!, [
              "role",
              "full_name",
              "phone",
              "employee_code",
              "hire_date",
              "shift",
            ]);
            setConfirmRole(false);
            return;
          }
          if (error.isCannotModifySelf()) {
            toast.error("You cannot change your own role.");
            setConfirmRole(false);
            return;
          }
          if (error.isInvalidRoleTransition()) {
            toast.error("This role change is not allowed.");
            setConfirmRole(false);
            return;
          }
          if (error.isEmployeeCodeExists()) {
            toast.error("That employee code is already in use.");
            setConfirmRole(false);
            return;
          }
          toast.error(error.message);
        },
      },
    );
  }

  return (
    <>
      <div className="max-w-2xl space-y-4">
        {user.role === USER_ROLE.admin || user.role === USER_ROLE.customer ? (
          <p className="text-sm text-amber-800 dark:text-amber-200">
            Current role: <strong>{user.role}</strong>. Select an operational role below
            (staff, baker, or manager). Platform admin accounts are created via dev seed
            only.
          </p>
        ) : null}
        <Form {...form}>
          <form
            className="space-y-4"
            noValidate
            onSubmit={form.handleSubmit(requestRoleSubmit)}
          >
            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <FormControl>
                    <select
                      className="h-10 w-full rounded-md border border-zinc-200 bg-white px-3 text-sm dark:border-zinc-800 dark:bg-zinc-950"
                      {...field}
                    >
                      {ASSIGNABLE_OPERATIONAL_ROLES.map((role) => (
                        <option key={role} value={role}>
                          {role}
                        </option>
                      ))}
                    </select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="full_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Full name (optional)</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="phone"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Phone</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      value={field.value ?? ""}
                      onChange={(e) => field.onChange(e.target.value || null)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="employee_code"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Employee code</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      value={field.value ?? ""}
                      onChange={(e) => field.onChange(e.target.value || null)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="hire_date"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Hire date</FormLabel>
                  <FormControl>
                    <Input
                      type="date"
                      {...field}
                      value={field.value ?? ""}
                      onChange={(e) => field.onChange(e.target.value || null)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="shift"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Shift</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      value={field.value ?? ""}
                      onChange={(e) => field.onChange(e.target.value || null)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button disabled={updateRole.isPending} type="submit">
              {updateRole.isPending ? "Updating…" : "Update role"}
            </Button>
          </form>
        </Form>
      </div>

      <ConfirmDialog
        confirmLabel="Apply role change"
        description="This may create or replace profile records based on the target role."
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

type AdminUserDetailRoleTabProps = {
  userId: string;
  user: AdminUser;
};

export function AdminUserDetailRoleTab({ userId, user }: AdminUserDetailRoleTabProps) {
  return (
    <AdminUserDetailRoleForm
      key={`${user.id}-${user.updated_at}`}
      userId={userId}
      user={user}
    />
  );
}
