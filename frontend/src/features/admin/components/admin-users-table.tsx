"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardSearchField } from "@/components/ui/dashboard-search-field";
import { DashboardTableActionLink } from "@/components/ui/dashboard-table-action-link";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTablePagePlaceholders } from "@/components/ui/dashboard-table-page-placeholders";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { Button } from "@/components/ui/button";
import {
  DASHBOARD_TABLE_PAGE_SIZE,
  dashboardTableEmptyFiltersMessage,
} from "@/constants/dashboard-table";
import { USER_ROLE, roleDisplayLabel } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { useDebouncedTableSearch } from "@/lib/hooks/use-debounced-table-search";
import { PAGE_TITLES } from "@/lib/metadata/page-title";
import { paginatedPlaceholderCountFromMeta } from "@/lib/pagination/dashboard-pagination";
import {
  isInitialQueryLoad,
  isQueryRefetching,
} from "@/lib/react-query/query-surface";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

import { useDisable, useEnable, useRevokeSessions, useUsers } from "../hooks";
import { adminUserListName } from "../lib/user-display";
import type { AdminUser, AdminUserRoleFilter } from "../schemas";
import { AdminUserAccountStatusToggle } from "./admin-user-account-status-toggle";
import { AdminUserSessionAction } from "./admin-user-session-action";

type PendingAction = "revoke" | "toggle";

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

  const revokeSessions = useRevokeSessions();
  const disableUser = useDisable();
  const enableUser = useEnable();
  const togglePending = disableUser.isPending || enableUser.isPending;
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
  const users = usersQuery.data?.users ?? [];
  const pagination = usersQuery.data?.pagination;
  const initialLoad = isInitialQueryLoad(usersQuery.isPending, usersQuery.data);
  const refetching = isQueryRefetching(
    usersQuery.isFetching,
    usersQuery.isPending,
    usersQuery.data,
  );
  const pagePlaceholderCount = paginatedPlaceholderCountFromMeta(
    users.length,
    pagination,
    PAGE_SIZE,
  );

  function clearPendingAction(): void {
    setPendingAction(null);
    setActionUser(null);
  }

  function handleActionError(error: unknown, fallback: string): void {
    if (isApiError(error) && error.isCannotModifySelf()) {
      toast.error("You cannot perform this action on your own account.");
      return;
    }
    toast.error(isApiError(error) ? error.message : fallback);
  }

  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        actions={
          <Link href={ROUTE.admin.usersNew}>
            <Button type="button">New user</Button>
          </Link>
        }
        description="Manage operational users and account status."
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

        <div className={cn("db-table-wrap", refetching && "is-refetching")}>
        <table className="db-table db-table--admin-users db-table--comfortable">
          <colgroup>
            <col className="db-table-col-email" />
            <col className="db-table-col-name" />
            <col className="db-table-col-role" />
            <col className="db-table-col-action" />
            <col className="db-table-col-revoke" />
            <col className="db-table-col-action" />
          </colgroup>
          <thead>
            <tr>
              <th className="db-table-cell-email">Email</th>
              <th className="db-table-cell-name">Name</th>
              <th>Role</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-revoke-sessions">Revoke sessions</th>
              <th className="db-table-detail">Detail</th>
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
              isInitialLoad={initialLoad}
            />
            {!initialLoad && !usersQuery.isError && users.length > 0
              ? users.map((user) => {
                const listName = adminUserListName(user);
                return (
                <tr key={user.id}>
                  <td
                    className="db-table-cell-email db-table-cell-primary db-table-cell-truncate"
                    title={user.email}
                  >
                    {user.email}
                  </td>
                  <td
                    className={cn(
                      "db-table-cell-name db-table-cell-truncate",
                      listName === "—"
                        ? "db-table-cell-placeholder"
                         : "text-muted",
                    )}
                    title={listName === "—" ? undefined : listName}
                  >
                    {listName}
                  </td>
                  <td className="db-table-cell-role">{roleDisplayLabel(user.role)}</td>
                  <td className="db-table-status">
                    <AdminUserAccountStatusToggle
                      currentUserId={currentUserId}
                      onToggleAccount={(target) => {
                        setActionUser(target);
                        setPendingAction("toggle");
                      }}
                      user={user}
                    />
                  </td>
                  <td className="db-table-revoke-sessions">
                    <AdminUserSessionAction
                      currentUserId={currentUserId}
                      onRevokeSessions={(target) => {
                        setActionUser(target);
                        setPendingAction("revoke");
                      }}
                      user={user}
                    />
                  </td>
                  <td className="db-table-detail">
                    <DashboardTableActionLink
                      href={ROUTE.admin.userDetail(user.id)}
                      label={`View ${user.email}`}
                    />
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
          itemLabel="users"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? users.length}
          totalPages={pagination?.total_pages ?? 1}
        />
        </div>
      </div>

      <ConfirmDialog
        confirmLabel="Revoke sessions"
        confirmVariant="warning"
        description="All active sessions for this user will be revoked."
        isPending={revokeSessions.isPending}
        onCancel={clearPendingAction}
        onConfirm={() => {
          if (!actionUser) {
            return;
          }
          revokeSessions.mutate(actionUser.id, {
            onSuccess: clearPendingAction,
            onError: (error) => {
              handleActionError(error, "Failed to revoke sessions");
              clearPendingAction();
            },
          });
        }}
        open={pendingAction === "revoke" && actionUser !== null}
        title="Revoke all sessions?"
      />

      <ConfirmDialog
        confirmLabel={
          actionUser?.disabled ? "Enable account" : "Disable account"
        }
        confirmVariant={actionUser?.disabled ? "default" : "destructive"}
        description={
          actionUser?.disabled
            ? "This account will be restored and can sign in again."
            : "This action soft-deletes the account and revokes active sessions."
        }
        isPending={togglePending}
        onCancel={clearPendingAction}
        onConfirm={() => {
          if (!actionUser) {
            return;
          }
          const mutation = actionUser.disabled ? enableUser : disableUser;
          mutation.mutate(actionUser.id, {
            onSuccess: clearPendingAction,
            onError: (error) => {
              handleActionError(
                error,
                actionUser.disabled
                  ? "Failed to enable user"
                  : "Failed to disable user",
              );
              clearPendingAction();
            },
          });
        }}
        open={pendingAction === "toggle" && actionUser !== null}
        title={
          actionUser?.disabled ? "Enable this account?" : "Disable this account?"
        }
      />
    </div>
  );
}
