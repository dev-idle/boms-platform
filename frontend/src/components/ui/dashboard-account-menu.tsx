"use client";

import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { LogOutIcon } from "@/components/icons/dashboard-nav-icons";
import { DashboardChevronUpIcon } from "@/components/icons/dashboard-ui-icons";
import { useLogout } from "@/features/auth";
import { userDisplayName, userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

type DashboardAccountMenuProps = {
  /** Account root for the signed-in role, e.g. `/admin/account`. */
  accountHref: string;
  profileHref: string;
};

/**
 * Everything that belongs to the person rather than the workspace: profile today,
 * appearance and language later. The sidebar nav stays a list of places to work.
 */
export function DashboardAccountMenu({
  accountHref,
  profileHref,
}: DashboardAccountMenuProps) {
  const pathname = usePathname();
  const user = useAuthStore((state) => state.user);
  const logout = useLogout();
  const displayName = userDisplayName(user);
  const email = user?.email ?? "";
  // No nav item owns the account routes any more, so the trigger carries "you are here".
  const onAccountRoute =
    pathname === accountHref || pathname.startsWith(`${accountHref}/`);

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        className={cn(
          "dashboard-account-trigger",
          onAccountRoute && "dashboard-account-trigger--current",
        )}
      >
        <span aria-hidden className="dashboard-avatar">
          {userInitials(user)}
        </span>
        <span className="dashboard-account-meta">
          <span className="dashboard-account-name" title={displayName}>
            {displayName}
          </span>
          {email ? (
            <span className="dashboard-account-email" title={email}>
              {email}
            </span>
          ) : null}
        </span>
        <DashboardChevronUpIcon className="dashboard-account-chevron" />
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="start"
          className="dashboard-account-popover"
          side="top"
          sideOffset={8}
        >
          <DropdownMenu.Label className="dashboard-account-popover__label">
            Account
          </DropdownMenu.Label>

          <DropdownMenu.Item asChild>
            <Link className="dashboard-account-item" href={profileHref}>
              Profile
            </Link>
          </DropdownMenu.Item>

          <DropdownMenu.Separator className="dashboard-account-separator" />

          {/* Selecting closes the menu and the mutation always redirects, so there is
              no pending state anyone can see. */}
          <DropdownMenu.Item
            className="dashboard-account-item dashboard-account-item--signout"
            onSelect={() => logout.mutate()}
          >
            Sign out
            {/* Trailing so both labels keep one left rail; the only glyph in the
                menu, so it reads as punctuation rather than a column of icons. */}
            <LogOutIcon className="dashboard-account-item__mark" />
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
