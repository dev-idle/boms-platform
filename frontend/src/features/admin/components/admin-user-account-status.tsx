import { USER_ROLE } from "@/constants/roles";
import { formatDateTime } from "@/lib/validation/datetime";
import { cn } from "@/lib/utils";

import type { AdminUser } from "../schemas";

type AdminUserAccountStatusProps = {
  user: AdminUser;
};

type ProfileFact = {
  label: string;
  /** Codes, phones and timestamps are data — mono. */
  mono?: boolean;
  value: string;
};

const PROFILE_TITLE_ID = "admin-user-profile-title";

function yesNo(value: boolean): string {
  return value ? "Yes" : "No";
}

function timestampFacts(user: AdminUser): ProfileFact[] {
  return [
    { label: "Created", mono: true, value: formatDateTime(user.created_at) },
    { label: "Updated", mono: true, value: formatDateTime(user.updated_at) },
  ];
}

function customerFacts(user: AdminUser): ProfileFact[] {
  return [
    { label: "Email", value: user.email },
    { label: "Email verified", value: yesNo(user.email_verified) },
    ...timestampFacts(user),
    { label: "Display name", value: user.display_name ?? "—" },
    { label: "Phone", mono: Boolean(user.phone), value: user.phone ?? "—" },
  ];
}

function internalFacts(user: AdminUser): ProfileFact[] {
  const facts: ProfileFact[] = [
    { label: "Email", value: user.email },
    { label: "Email verified", value: yesNo(user.email_verified) },
    { label: "Must change password", value: yesNo(user.must_change_password) },
  ];

  if (user.role !== USER_ROLE.admin) {
    facts.push({
      label: "Employee code",
      mono: Boolean(user.employee_code),
      value: user.employee_code ?? "—",
    });
  }

  return [
    ...facts,
    { label: "Full name", value: user.full_name ?? "—" },
    { label: "Phone", mono: Boolean(user.phone), value: user.phone ?? "—" },
    ...timestampFacts(user),
  ];
}

/**
 * Read-only account facts as the detail page's sticky rail (archetype C): facts
 * stay visible while the admin works the role form and activity log.
 */
export function AdminUserAccountStatus({ user }: AdminUserAccountStatusProps) {
  const facts =
    user.role === USER_ROLE.customer ? customerFacts(user) : internalFacts(user);

  return (
    <aside aria-labelledby={PROFILE_TITLE_ID} className="admin-user-profile">
      <h2 className="admin-user-profile__title" id={PROFILE_TITLE_ID}>
        Profile
      </h2>
      <dl className="admin-user-profile__facts" id="admin-user-profile">
        {facts.map((fact) => (
          <div key={fact.label}>
            <dt>{fact.label}</dt>
            <dd className={cn(fact.mono && "admin-user-profile__value--mono")}>
              {fact.value}
            </dd>
          </div>
        ))}
      </dl>
    </aside>
  );
}
