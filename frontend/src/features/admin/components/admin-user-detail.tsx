"use client";

import Link from "next/link";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";

import { useUserDetail } from "../hooks";

import { AdminUserDetailProfileTab } from "./admin-user-detail-profile-tab";
import { AdminUserDetailRoleTab } from "./admin-user-detail-role-tab";
import { AdminUserDetailSessionsTab } from "./admin-user-detail-sessions-tab";

type Tab = "profile" | "role" | "sessions";

export function AdminUserDetail({ userId }: { userId: string }) {
  const [tab, setTab] = useState<Tab>("profile");
  const userQuery = useUserDetail(userId);
  const user = userQuery.data;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-medium text-ink">
            User detail
          </h1>
          {user ? (
            <p className="mt-2 text-sm text-ink-2">
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
        <p className="text-sm text-muted">Loading user details…</p>
      ) : !user ? (
        <p className="text-sm text-muted">Unable to load this user.</p>
      ) : null}

      {tab === "profile" && user ? (
        <AdminUserDetailProfileTab userId={userId} user={user} />
      ) : null}

      {tab === "role" && user ? (
        <AdminUserDetailRoleTab userId={userId} user={user} />
      ) : null}

      {tab === "sessions" && user ? (
        <AdminUserDetailSessionsTab userId={userId} />
      ) : null}
    </div>
  );
}
