"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { USER_ROLE } from "@/constants/roles";
import { Button } from "@/components/ui/button";
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

import { useUpdateUserProfile } from "../hooks";
import { applyFormFieldErrors } from "../lib/apply-form-validation";
import { adminUserToProfileFormValues } from "../lib/map-user-form-values";
import {
  adminUserProfileFormSchema,
  type AdminUser,
  type AdminUserProfileFormValues,
} from "../schemas";

type AdminUserDetailProfileFormProps = {
  userId: string;
  user: AdminUser;
};

function AdminUserDetailProfileForm({
  userId,
  user,
}: AdminUserDetailProfileFormProps) {
  const updateProfile = useUpdateUserProfile();

  const form = useForm<AdminUserProfileFormValues>({
    resolver: zodResolver(adminUserProfileFormSchema),
    defaultValues: adminUserToProfileFormValues(user),
  });

  function onSubmit(values: AdminUserProfileFormValues): void {
    updateProfile.mutate(
      {
        id: userId,
        input: {
          full_name: values.full_name.trim(),
          phone: values.phone?.trim() ? values.phone.trim() : null,
          employee_code: values.employee_code?.trim()
            ? values.employee_code.trim()
            : null,
        },
      },
      {
        onError: (error) => {
          if (!isApiError(error)) {
            toast.error("Failed to update profile");
            return;
          }
          if (error.hasValidationDetails()) {
            applyFormFieldErrors(form, error.details!, [
              "full_name",
              "phone",
              "employee_code",
            ]);
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

  const isStaffLike =
    user.role === USER_ROLE.staff ||
    user.role === USER_ROLE.baker ||
    user.role === USER_ROLE.manager;

  return (
    <div className="max-w-2xl">
      <Form {...form}>
        <form
          className="space-y-4"
          noValidate
          onSubmit={form.handleSubmit(onSubmit)}
        >
          <FormField
            control={form.control}
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
            control={form.control}
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

          {isStaffLike ? (
            <>
              <FormField
                control={form.control}
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

            </>
          ) : null}

          <Button disabled={updateProfile.isPending} type="submit">
            {updateProfile.isPending ? "Saving…" : "Save profile"}
          </Button>
        </form>
      </Form>
    </div>
  );
}

type AdminUserDetailProfileTabProps = {
  userId: string;
  user: AdminUser;
};

export function AdminUserDetailProfileTab({
  userId,
  user,
}: AdminUserDetailProfileTabProps) {
  if (user.role === USER_ROLE.customer) {
    return (
      <p className="text-sm text-ink-2">
        Customer profiles are not editable from this endpoint.
      </p>
    );
  }

  return (
    <AdminUserDetailProfileForm
      key={`${user.id}-${user.updated_at}`}
      userId={userId}
      user={user}
    />
  );
}
