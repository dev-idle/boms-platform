"use client";

import type { UseMutationResult } from "@tanstack/react-query";
import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTablePagePlaceholders } from "@/components/ui/dashboard-table-page-placeholders";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { StatusPill } from "@/components/ui/status-pill";
import { Button } from "@/components/ui/button";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import {
  DASHBOARD_TABLE_PAGE_SIZE,
  dashboardTableEmptyFiltersMessage,
} from "@/constants/dashboard-table";
import { USER_ROLE, roleDisplayLabel } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { getDashboardQuerySurface } from "@/lib/react-query/query-surface";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

import { useDisable, useEnable, useRevokeSessions, useUsers } from "../hooks";
import {
  adminUserAccountStatus,
  adminUserRowActions,
} from "../lib/account-management";
import { adminUserInitials, adminUserListName } from "../lib/user-display";
import type { AdminUser, AdminUserRoleFilter } from "../schemas";

type PendingAction = "disable" | "enable" | "revoke";

const PAGE_SIZE = DASHBOARD_TABLE_PAGE_SIZE;

const ROLE_FILTERS: Array<{ value: AdminUserRoleFilter | undefined; label: string }> = [
  { value: undefined, label: "All roles" },
  { value: USER_ROLE.customer, label: roleDisplayLabel(USER_ROLE.customer) },
  { value: USER_ROLE.staff, label: roleDisplayLabel(USER_ROLE.staff) },
  { value: USER_ROLE.baker, label: roleDisplayLabel(USER_ROLE.baker) },
  { value: USER_ROLE.manager, label: roleDisplayLabel(USER_ROLE.manager) },
  { value: USER_ROLE.admin, label: roleDisplayLabel(USER_ROLE.admin) },
];

export function AdminUsersTable() {
  const currentUserId = useAuthStore((state) => state.user?.id);
  const [role, setRole] = useState<AdminUserRoleFilter | undefined>(undefined);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);
  const [actionUser, setActionUser] = useState<AdminUser | null>(null);

  const disableUser = useDisable();
  const enableUser = useEnable();
  const revokeSessions = useRevokeSessions();
  const {
    clear,
    input,
    page,
    search,
    setInput,
    setPage,
  } = useDebouncedTableSearch();

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search, role }),
    [page, role, search],
  );
  const usersQuery = useUsers(filter);
  const { initialLoading, refetching } = getDashboardQuerySurface(usersQuery);
  const users = usersQuery.data?.users ?? [];
  const pagination = usersQuery.data?.pagination;
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    users.length,
    pagination,
    PAGE_SIZE,
  );

  function requestAction(action: PendingAction, target: AdminUser): void {
    setActionUser(target);
    setPendingAction(action);
  }

  function clearPendingAction(): void {
    setPendingAction(null);
    setActionUser(null);
  }

  function handleActionError(error: unknown, fallback: string): void {
    if (isApiError(error) && error.isCannotModifySelf()) {
      toast.error("You cannot perform this action on your own account.");
      return;
    }
    if (isApiError(error) && error.isCannotModifyAdmin()) {
      toast.error("Admin accounts cannot be disabled or have sessions revoked.");
      return;
    }
    toast.error(isApiError(error) ? error.message : fallback);
  }

  function runPendingAction<TData>(
    mutation: UseMutationResult<TData, Error, string>,
    fallback: string,
  ): void {
    if (!actionUser) {
      return;
    }
    mutation.mutate(actionUser.id, {
      onSuccess: clearPendingAction,
      onError: (error) => {
        handleActionError(error, fallback);
        clearPendingAction();
      },
    });
  }

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Button asChild>
            <Link href={ROUTE.admin.usersNew}>+ New user</Link>
          </Button>
        }
        description="Manage operational users and account status."
        eyebrow={DASHBOARD_PAGE_EYEBROW.accountsAccess}
        title={PAGE_TITLES.users}
      />

      <div className="dashboard-page-body">
        <div className="db-table-filters">
          <DashboardSearchField
            onChange={setInput}
            onClear={clear}
            placeholder="Search email, name, employee code"
            value={input}
          />
          <DashboardFilterGroup
            aria-label="Filter by role"
            onChange={(next) => {
              setRole(next);
              setPage(1);
            }}
            options={ROLE_FILTERS}
            value={role}
          />
        </div>

        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--admin-users">
            <colgroup>
              <col className="db-table-col-avatar" />
              <col className="db-table-col-email" />
              <col className="db-table-col-name" />
              <col className="db-table-col-role" />
              <col className="db-table-col-status" />
              <col className="db-table-col-actions" />
            </colgroup>
            <thead>
              <tr>
                <th className="db-table-cell-avatar">
                  <span className="sr-only">Avatar</span>
                </th>
                <th className="db-table-cell-email">Email</th>
                <th className="db-table-cell-name">Name</th>
                <th>Role</th>
                <th className="db-table-status">Status</th>
                <th className="db-table-detail">Actions</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={6}
                emptyFilteredMessage={dashboardTableEmptyFiltersMessage("users")}
                entityLabel="users"
                hasActiveFilter={Boolean(search || role)}
                isEmpty={users.length === 0}
                isError={usersQuery.isError}
                initialLoading={initialLoading}
              />
              {!initialLoading && !usersQuery.isError && users.length > 0
                ? users.map((user) => {
                    const listName = adminUserListName(user);
                    const accountStatus = adminUserAccountStatus(user);
                    const { account, sessions } = adminUserRowActions(user, currentUserId);
                    const accountAction = account.label === "Enable" ? "enable" : "disable";

                    return (
                      <tr key={user.id}>
                        <td className="db-table-cell-avatar">
                          <span aria-hidden className="db-table-avatar">
                            {adminUserInitials(user)}
                          </span>
                        </td>
                        <td
                          className="db-table-cell-email db-table-cell-primary db-table-cell-truncate"
                          title={user.email}
                        >
                          {user.email}
                        </td>
                        <td
                          className={cn(
                            "db-table-cell-name db-table-cell-truncate",
                            listName === "—" ? "db-table-cell-placeholder" : "text-muted",
                          )}
                          title={listName === "—" ? undefined : listName}
                        >
                          {listName}
                        </td>
                        <td className="db-table-cell-role">{roleDisplayLabel(user.role)}</td>
                        <td className="db-table-status">
                          <StatusPill
                            label={accountStatus.label}
                            variant={accountStatus.variant}
                          />
                        </td>
                        <td className="db-table-detail">
                          {/* Fixed slots — destructive, cautionary, navigational. A slot
                              that cannot apply is greyed in place, never removed. */}
                          <DashboardTableRowActions>
                            <DashboardTableActionButton
                              blockedReason={
                                account.kind === "blocked" ? account.reason : undefined
                              }
                              label={`${account.label} account for ${user.email}`}
                              onClick={() => requestAction(accountAction, user)}
                              text={account.label}
                              tone={accountAction === "enable" ? "accent" : "danger"}
                            />
                            <DashboardTableActionButton
                              blockedReason={
                                sessions.kind === "blocked" ? sessions.reason : undefined
                              }
                              label={`Revoke sessions for ${user.email}`}
                              onClick={() => requestAction("revoke", user)}
                              text="Revoke"
                              tone="warning"
                            />
                            <DashboardTableActionLink
                              href={ROUTE.admin.userDetail(user.id)}
                              label={`View ${user.email}`}
                              text="Detail"
                            />
                          </DashboardTableRowActions>
                        </td>
                      </tr>
                    );
                  })
                : null}
              <DashboardTablePagePlaceholders
                columnCount={6}
                count={pagePlaceholderCount}
              />
            </tbody>
          </table>
          <DashboardTablePagination
            disabled={usersQuery.isFetching}
            itemLabel="accounts"
            onPageChange={setPage}
            page={pagination?.page ?? page}
            pageSize={pagination?.page_size ?? PAGE_SIZE}
            totalItems={pagination?.total ?? users.length}
            totalPages={pagination?.total_pages ?? 1}
          />
        </DashboardTableWrap>
      </div>

      <ConfirmDialog
        confirmLabel="Disable account"
        confirmVariant="destructive"
        description="This action soft-deletes the account and revokes active sessions."
        isPending={disableUser.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => runPendingAction(disableUser, "Failed to disable account")}
        open={pendingAction === "disable" && actionUser !== null}
        title="Disable account?"
      />

      <ConfirmDialog
        confirmLabel="Enable account"
        description="This account will be restored and can sign in again."
        isPending={enableUser.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => runPendingAction(enableUser, "Failed to enable account")}
        open={pendingAction === "enable" && actionUser !== null}
        title="Enable account?"
      />

      <ConfirmDialog
        confirmLabel="Revoke sessions"
        confirmVariant="warning"
        description="All active sessions for this user will be revoked."
        isPending={revokeSessions.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => runPendingAction(revokeSessions, "Failed to revoke sessions")}
        open={pendingAction === "revoke" && actionUser !== null}
        title="Revoke all sessions?"
      />
    </div>
  );
}
