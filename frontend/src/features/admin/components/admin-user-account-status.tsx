import { USER_ROLE } from "@/constants/roles";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatVietnamPhone } from "@/lib/validation/phone";
import { cn } from "@/lib/utils";

import type { AdminUser } from "../schemas";

type AdminUserAccountStatusProps = {
  user: AdminUser;
};

type ProfileFact = {
  label: string;
  /** Codes, phones and timestamps are data — mono. */
  mono?: boolean;
  /** First fact of a group; takes the hairline that opens it. */
  startsGroup?: boolean;
  value: string;
};

const PROFILE_TITLE_ID = "admin-user-profile-title";

function yesNo(value: boolean): string {
  return value ? "Yes" : "No";
}

function dash(value: string | null | undefined): string {
  return value ?? "—";
}

/**
 * Three groups in one order for every role: who they are, what the account is,
 * when it changed.
 *
 * Employee code belongs to the first group, straight after the email: it is an
 * identifier assigned once and kept for the life of the account, not a flag like
 * the two that follow. It is not the first row even for the roles that have one,
 * because admins and customers have none — leading with it would give the panel a
 * different opening row per role, which is the inconsistency it is meant to fix.
 *
 * Role is deliberately absent. The page header prints it and the role form owns
 * it, and a read-only third copy would sit visibly stale the moment an admin
 * changes that select without saving.
 */
function profileFacts(user: AdminUser): ProfileFact[] {
  const isCustomer = user.role === USER_ROLE.customer;
  const hasEmployeeCode = !isCustomer && user.role !== USER_ROLE.admin;

  const identity: ProfileFact[] = [{ label: "Email", value: user.email }];
  if (hasEmployeeCode) {
    identity.push({
      label: "Employee code",
      mono: Boolean(user.employee_code),
      value: dash(user.employee_code),
    });
  }
  identity.push(
    {
      label: isCustomer ? "Display name" : "Full name",
      value: dash(isCustomer ? user.display_name : user.full_name),
    },
    {
      label: "Phone",
      mono: Boolean(user.phone),
      value: formatVietnamPhone(user.phone) || "—",
    },
  );

  const account: ProfileFact[] = [
    { label: "Email verified", startsGroup: true, value: yesNo(user.email_verified) },
  ];
  if (!isCustomer) {
    account.push({
      label: "Must change password",
      value: yesNo(user.must_change_password),
    });
  }

  const audit: ProfileFact[] = [
    {
      label: "Created",
      mono: true,
      startsGroup: true,
      value: formatDateTime(user.created_at),
    },
    { label: "Updated", mono: true, value: formatDateTime(user.updated_at) },
  ];

  return [...identity, ...account, ...audit];
}

/**
 * Read-only account facts as the detail page's sticky rail (archetype C): facts
 * stay visible while the admin works the role form and activity log.
 */
export function AdminUserAccountStatus({ user }: AdminUserAccountStatusProps) {
  return (
    <aside aria-labelledby={PROFILE_TITLE_ID} className="admin-user-profile">
      <h2 className="admin-user-profile__title" id={PROFILE_TITLE_ID}>
        Profile
      </h2>
      <dl className="admin-user-profile__facts" id="admin-user-profile">
        {profileFacts(user).map((fact) => (
          <div
            className={cn(fact.startsGroup && "admin-user-profile__group")}
            key={fact.label}
          >
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
