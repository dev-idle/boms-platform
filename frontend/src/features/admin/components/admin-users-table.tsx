"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ROUTE } from "@/constants/routes";

import { useUsers } from "../hooks";
import type { AdminUser } from "../schemas";

const PAGE_SIZE = 20;

export function AdminUsersTable() {
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);

  const filter = useMemo(
    () => ({ page, page_size: PAGE_SIZE, search }),
    [page, search],
  );
  const usersQuery = useUsers(filter);

  const users = usersQuery.data?.users ?? [];
  const pagination = usersQuery.data?.pagination;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-page-title">
            Admin users
          </h1>
          <p className="mt-2 text-sm text-ink-2">
            Manage operational users and account status.
          </p>
        </div>
        <Link href={ROUTE.admin.usersNew}>
          <Button type="button">New user</Button>
        </Link>
      </div>

      <form
        className="flex flex-wrap items-center gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          setPage(1);
          setSearch(searchInput.trim());
        }}
      >
        <Input
          className="max-w-sm"
          onChange={(event) => setSearchInput(event.target.value)}
          placeholder="Search email, name, employee code"
          value={searchInput}
        />
        <Button type="submit">Search</Button>
        {search ? (
          <Button
            onClick={() => {
              setSearchInput("");
              setSearch("");
              setPage(1);
            }}
            type="button"
            variant="outline"
          >
            Clear
          </Button>
        ) : null}
      </form>

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="min-w-full divide-y divide-border text-sm">
          <thead className="bg-surface-alt">
            <tr className="text-left text-xs uppercase tracking-wide text-muted">
              <th className="px-4 py-3">Email</th>
              <th className="px-4 py-3">Role</th>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {usersQuery.isPending ? (
              <tr>
                <td className="px-4 py-4 text-muted" colSpan={5}>
                  Loading users…
                </td>
              </tr>
            ) : usersQuery.isError ? (
              <tr>
                <td className="px-4 py-4 text-error" colSpan={5}>
                  Failed to load users.
                </td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td className="px-4 py-4 text-muted" colSpan={5}>
                  No users found.
                </td>
              </tr>
            ) : (
              users.map((user: AdminUser) => (
                <tr key={user.id}>
                  <td className="px-4 py-3">{user.email}</td>
                  <td className="px-4 py-3 uppercase">{user.role}</td>
                  <td className="px-4 py-3">
                    {user.full_name ?? user.display_name ?? "-"}
                  </td>
                  <td className="px-4 py-3">
                    {user.disabled ? "Disabled" : "Active"}
                  </td>
                  <td className="px-4 py-3">
                    <Link
                      className="text-sm font-medium text-ink underline"
                      href={ROUTE.admin.userDetail(user.id)}
                    >
                      Detail
                    </Link>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-sm text-ink-2">
          Page {pagination?.page ?? page} of {pagination?.total_pages ?? 1}
        </p>
        <div className="flex gap-2">
          <Button
            disabled={page <= 1 || usersQuery.isFetching}
            onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
            type="button"
            variant="outline"
          >
            Previous
          </Button>
          <Button
            disabled={
              (!!pagination && page >= pagination.total_pages) ||
              usersQuery.isFetching
            }
            onClick={() => setPage((prev) => prev + 1)}
            type="button"
            variant="outline"
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
