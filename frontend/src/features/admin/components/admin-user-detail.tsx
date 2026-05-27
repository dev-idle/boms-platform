"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { toast } from "sonner";

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
import { ASSIGNABLE_OPERATIONAL_ROLES, USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import {
  useDisable,
  useRevokeSessions,
  useUpdateRole,
  useUpdateUserProfile,
  useUserDetail,
} from "../hooks";
import { updateRoleSchema, type UpdateRoleInput } from "../schemas";

type Tab = "profile" | "role" | "sessions";

const profileFormSchema = z.object({
  full_name: z.string().trim().min(1, "Full name is required").max(255),
  phone: z.string().trim().max(50).optional(),
  employee_code: z.string().trim().max(64).optional(),
  hire_date: z.string().trim().optional(),
  shift: z.string().trim().max(64).optional(),
});

type ProfileFormValues = z.infer<typeof profileFormSchema>;

type RoleFormValues = UpdateRoleInput;

export function AdminUserDetail({ userId }: { userId: string }) {
  const [tab, setTab] = useState<Tab>("profile");
  const [confirmRole, setConfirmRole] = useState(false);
  const [pendingRoleValues, setPendingRoleValues] = useState<RoleFormValues | null>(
    null,
  );
  const [confirmDisable, setConfirmDisable] = useState(false);
  const [confirmRevoke, setConfirmRevoke] = useState(false);

  const userQuery = useUserDetail(userId);
  const updateProfile = useUpdateUserProfile();
  const updateRole = useUpdateRole();
  const disableUser = useDisable();
  const revokeSessions = useRevokeSessions();

  const profileForm = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      full_name: "",
      phone: "",
      employee_code: "",
      hire_date: "",
      shift: "",
    },
  });

  const roleForm = useForm<RoleFormValues>({
    resolver: zodResolver(updateRoleSchema),
    defaultValues: {
      role: USER_ROLE.staff,
      full_name: "",
      phone: "",
      employee_code: "",
      hire_date: "",
      shift: "",
    },
  });

  useEffect(() => {
    const user = userQuery.data;
    if (!user) {
      return;
    }
    profileForm.reset({
      full_name: user.full_name ?? "",
      phone: user.phone ?? "",
      employee_code: user.employee_code ?? "",
      hire_date: user.hire_date ? user.hire_date.slice(0, 10) : "",
      shift: user.shift ?? "",
    });
    const roleForForm =
      user.role === USER_ROLE.staff ||
      user.role === USER_ROLE.baker ||
      user.role === USER_ROLE.manager
        ? user.role
        : USER_ROLE.staff;
    roleForm.reset({
      role: roleForForm,
      full_name: user.full_name ?? "",
      phone: user.phone ?? "",
      employee_code: user.employee_code ?? "",
      hire_date: user.hire_date ? user.hire_date.slice(0, 10) : "",
      shift: user.shift ?? "",
    });
  }, [profileForm, roleForm, userQuery.data]);

  const user = userQuery.data;
  const canEditOperationalProfile = user?.role !== "customer";

  function mapApiDetailsToForm(
    details: Record<string, string>,
    assign: (field: string, message: string) => void,
  ): void {
    for (const item of mapValidationDetailsToFormErrors(details)) {
      assign(item.field, item.message);
    }
  }

  function handleProfileSubmit(values: ProfileFormValues): void {
    updateProfile.mutate(
      {
        id: userId,
        input: {
          full_name: values.full_name.trim(),
          phone: values.phone?.trim() ? values.phone.trim() : null,
          employee_code: values.employee_code?.trim()
            ? values.employee_code.trim()
            : null,
          hire_date: values.hire_date?.trim() ? values.hire_date.trim() : null,
          shift: values.shift?.trim() ? values.shift.trim() : null,
        },
      },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Failed to update profile");
            return;
          }
          if (error.hasValidationDetails()) {
            mapApiDetailsToForm(error.details!, (field, message) => {
              if (
                field === "full_name" ||
                field === "phone" ||
                field === "employee_code" ||
                field === "hire_date" ||
                field === "shift"
              ) {
                profileForm.setError(field, { message });
              }
            });
            return;
          }
          if (error.isCannotModifySelf()) {
            toast.error("You cannot modify your own account from this screen.");
            return;
          }
          toast.error(error.message);
        },
      },
    );
  }

  function requestRoleSubmit(values: RoleFormValues): void {
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
          if (error.hasValidationDetails()) {
            mapApiDetailsToForm(error.details!, (field, message) => {
              if (field in pendingRoleValues) {
                roleForm.setError(field as keyof RoleFormValues, { message });
              }
            });
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">
            User detail
          </h1>
          {user ? (
            <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-300">
              {user.email} · {user.role}
            </p>
          ) : null}
        </div>
        <Link href={ROUTE.admin.users}>
          <Button type="button" variant="outline">
            Back to users
          </Button>
        </Link>
      </div>

      <div className="flex gap-2">
        <Button
          onClick={() => setTab("profile")}
          type="button"
          variant={tab === "profile" ? "default" : "outline"}
        >
          Profile
        </Button>
        <Button
          onClick={() => setTab("role")}
          type="button"
          variant={tab === "role" ? "default" : "outline"}
        >
          Role
        </Button>
        <Button
          onClick={() => setTab("sessions")}
          type="button"
          variant={tab === "sessions" ? "default" : "outline"}
        >
          Sessions
        </Button>
      </div>

      {userQuery.isPending ? (
        <p className="text-sm text-zinc-500">Loading user details…</p>
      ) : !user ? (
        <p className="text-sm text-zinc-500">Unable to load this user.</p>
      ) : null}

      {tab === "profile" && user ? (
        <div className="max-w-2xl">
          {canEditOperationalProfile ? (
            <Form {...profileForm}>
              <form
                className="space-y-4"
                noValidate
                onSubmit={profileForm.handleSubmit(handleProfileSubmit)}
              >
                <FormField
                  control={profileForm.control}
                  name="full_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Full name</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={profileForm.control}
                  name="phone"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Phone</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {(user.role === "staff" ||
                  user.role === "baker" ||
                  user.role === "manager") && (
                  <>
                    <FormField
                      control={profileForm.control}
                      name="employee_code"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Employee code</FormLabel>
                          <FormControl>
                            <Input {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={profileForm.control}
                      name="hire_date"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Hire date</FormLabel>
                          <FormControl>
                            <Input type="date" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <FormField
                      control={profileForm.control}
                      name="shift"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Shift</FormLabel>
                          <FormControl>
                            <Input {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </>
                )}

                <Button disabled={updateProfile.isPending} type="submit">
                  {updateProfile.isPending ? "Saving…" : "Save profile"}
                </Button>
              </form>
            </Form>
          ) : (
            <p className="text-sm text-zinc-600 dark:text-zinc-300">
              Customer profiles are not editable from this endpoint.
            </p>
          )}
        </div>
      ) : null}

      {tab === "role" && user ? (
        <div className="max-w-2xl space-y-4">
          {user.role === USER_ROLE.admin || user.role === USER_ROLE.customer ? (
            <p className="text-sm text-amber-800 dark:text-amber-200">
              Current role: <strong>{user.role}</strong>. Select an operational role below
              (staff, baker, or manager). Platform admin accounts are created via dev seed only.
            </p>
          ) : null}
          <Form {...roleForm}>
            <form
              className="space-y-4"
              noValidate
              onSubmit={roleForm.handleSubmit(requestRoleSubmit)}
            >
              <FormField
                control={roleForm.control}
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
                control={roleForm.control}
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
                control={roleForm.control}
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
                control={roleForm.control}
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
                control={roleForm.control}
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
                control={roleForm.control}
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
      ) : null}

      {tab === "sessions" && user ? (
        <div className="max-w-2xl space-y-4">
          <div className="rounded-lg border border-zinc-200 p-4 dark:border-zinc-800">
            <h2 className="font-semibold text-zinc-900 dark:text-zinc-100">
              Revoke sessions
            </h2>
            <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-300">
              Force this user to sign in again on all devices.
            </p>
            <Button
              className="mt-4"
              onClick={() => setConfirmRevoke(true)}
              type="button"
              variant="outline"
            >
              Revoke sessions
            </Button>
          </div>

          <div className="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-900/60 dark:bg-red-950/40">
            <h2 className="font-semibold text-red-800 dark:text-red-300">
              Disable user
            </h2>
            <p className="mt-1 text-sm text-red-700 dark:text-red-200">
              This action soft-deletes the account.
            </p>
            <Button
              className="mt-4"
              onClick={() => setConfirmDisable(true)}
              type="button"
              variant="outline"
            >
              Disable account
            </Button>
          </div>
        </div>
      ) : null}

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

      <ConfirmDialog
        confirmLabel="Disable user"
        description="The user account will be disabled immediately."
        isPending={disableUser.isPending}
        onCancel={() => setConfirmDisable(false)}
        onConfirm={() =>
          disableUser.mutate(userId, {
            onSuccess: () => {
              setConfirmDisable(false);
            },
            onError: (error) => {
              if (!isApiError(error)) {
                toast.error("Failed to disable user");
                return;
              }
              if (error.isCannotModifySelf()) {
                toast.error("You cannot disable your own account.");
                return;
              }
              toast.error(error.message);
            },
          })
        }
        open={confirmDisable}
        title="Disable this user?"
      />

      <ConfirmDialog
        confirmLabel="Revoke sessions"
        description="All active sessions for this user will be revoked."
        isPending={revokeSessions.isPending}
        onCancel={() => setConfirmRevoke(false)}
        onConfirm={() =>
          revokeSessions.mutate(userId, {
            onSuccess: () => {
              setConfirmRevoke(false);
            },
            onError: (error) => {
              if (!isApiError(error)) {
                toast.error("Failed to revoke sessions");
                return;
              }
              if (error.isCannotModifySelf()) {
                toast.error("Use logout to end your own sessions.");
                return;
              }
              toast.error(error.message);
            },
          })
        }
        open={confirmRevoke}
        title="Revoke all sessions?"
      />
    </div>
  );
}
