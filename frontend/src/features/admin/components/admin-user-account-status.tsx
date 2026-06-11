import type { ReactNode } from "react";

import { USER_ROLE } from "@/constants/roles";
import { DashboardProfileSection } from "@/features/user";
import { formatDateTime } from "@/lib/validation/datetime";

import type { AdminUser } from "../schemas";

type AdminUserAccountStatusProps = {
  user: AdminUser;
};

type StatusFieldConfig = {
  label: string;
  value: ReactNode;
};

function StatusField({
  label,
  value,
}: {
  label: string;
  value: ReactNode;
}) {
  return (
    <div className="admin-profile-field">
      <p className="admin-profile-field-label">{label}</p>
      <div className="admin-profile-field-value">{value}</div>
    </div>
  );
}

function sharedAccountFields(user: AdminUser): StatusFieldConfig[] {
  return [
    { label: "Email", value: user.email },
    {
      label: "Email verified",
      value: user.email_verified ? "Yes" : "No",
    },
  ];
}

function internalAccountFields(user: AdminUser): StatusFieldConfig[] {
  return [
    ...sharedAccountFields(user),
    {
      label: "Must change password",
      value: user.must_change_password ? "Yes" : "No",
    },
    { label: "Created", value: formatDateTime(user.created_at) },
    { label: "Updated", value: formatDateTime(user.updated_at) },
  ];
}

function customerAccountFields(user: AdminUser): StatusFieldConfig[] {
  return [
    ...sharedAccountFields(user),
    { label: "Created", value: formatDateTime(user.created_at) },
    { label: "Updated", value: formatDateTime(user.updated_at) },
  ];
}

function customerProfileFields(user: AdminUser): StatusFieldConfig[] {
  return [
    { label: "Display name", value: user.display_name ?? "—" },
    { label: "Phone", value: user.phone ?? "—" },
  ];
}

function internalProfileFields(user: AdminUser): StatusFieldConfig[] {
  const fields: StatusFieldConfig[] = [
    { label: "Full name", value: user.full_name ?? "—" },
    { label: "Phone", value: user.phone ?? "—" },
  ];

  if (
    user.role === USER_ROLE.staff ||
    user.role === USER_ROLE.baker ||
    user.role === USER_ROLE.manager
  ) {
    fields.push({
      label: "Employee code",
      value: user.employee_code ?? "—",
    });
  }

  return fields;
}

function fieldsForUser(user: AdminUser): StatusFieldConfig[] {
  if (user.role === USER_ROLE.customer) {
    return [...customerAccountFields(user), ...customerProfileFields(user)];
  }

  return [...internalAccountFields(user), ...internalProfileFields(user)];
}

export function AdminUserAccountStatus({ user }: AdminUserAccountStatusProps) {
  const fields = fieldsForUser(user);

  return (
    <DashboardProfileSection id="admin-user-profile" title="Profile">
      <div className="admin-profile-field-grid">
        {fields.map((field) => (
          <StatusField key={field.label} label={field.label} value={field.value} />
        ))}
      </div>
    </DashboardProfileSection>
  );
}
