"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

import {
  createOperational,
  disableUser,
  enableUser,
  getUserById,
  listUsers,
  listUserActivity,
  resetUserPassword,
  revokeUserSessions,
  updateRole,
} from "../api";
import {
  listFilterSchema,
  userActivityFilterSchema,
  type AdminUser,
  type AdminResetPasswordResponse,
  type CreateOperationalInput,
  type CreateOperationalResponse,
  type ListFilterInput,
  type UpdateRoleInput,
  type UserActivityFilterInput,
} from "../schemas";
import { adminQueryKeys } from "./query-options";

export { adminQueryKeys } from "./query-options";

function normalizeFilter(input: ListFilterInput): ListFilterInput {
  return listFilterSchema.parse(input);
}

function hydrateUserCaches(queryClient: ReturnType<typeof useQueryClient>, user: AdminUser): void {
  queryClient.setQueryData(adminQueryKeys.user(user.id), user);
}

function keepAdminUsersPageData<T>(
  previousData: T | undefined,
  previousQuery: { queryKey: readonly unknown[] } | undefined,
  filter: ListFilterInput,
): T | undefined {
  const previousKey = previousQuery?.queryKey;
  const nextKey = adminQueryKeys.users(filter);
  if (!previousKey || !previousData) {
    return undefined;
  }

  const sameFilters =
    previousKey[3] === nextKey[3] &&
    previousKey[4] === nextKey[4] &&
    previousKey[5] === nextKey[5];

  return sameFilters ? previousData : undefined;
}

export function useUsers(input: ListFilterInput) {
  const filter = normalizeFilter(input);
  return useQuery({
    queryKey: adminQueryKeys.users(filter),
    queryFn: () => listUsers(filter),
    placeholderData: (previousData, previousQuery) =>
      keepAdminUsersPageData(previousData, previousQuery, filter),
  });
}

export function useUserDetail(id: string) {
  return useQuery({
    queryKey: adminQueryKeys.user(id),
    queryFn: () => getUserById(id),
    enabled: Boolean(id),
  });
}

export function useCreateOperational() {
  const queryClient = useQueryClient();
  const [tempPasswordData, setTempPasswordData] =
    useState<CreateOperationalResponse | null>(null);

  const mutation = useMutation({
    mutationFn: (input: CreateOperationalInput) => createOperational(input),
    onSuccess: (data) => {
      setTempPasswordData(data);
      hydrateUserCaches(queryClient, data.user);
      void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
      toast.success("User created");
    },
  });

  return {
    ...mutation,
    tempPasswordData,
    clearTempPasswordData: () => setTempPasswordData(null),
  };
}

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: { id: string; input: UpdateRoleInput }) =>
      updateRole(params.id, params.input),
    onSuccess: (user) => {
      hydrateUserCaches(queryClient, user);
      invalidateUserDetailCaches(queryClient, user.id);
      toast.success("Role updated");
    },
  });
}

function invalidateUserDetailCaches(
  queryClient: ReturnType<typeof useQueryClient>,
  userId: string,
): void {
  void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
  void queryClient.invalidateQueries({ queryKey: adminQueryKeys.user(userId) });
  void queryClient.invalidateQueries({
    queryKey: adminQueryKeys.userActivityRoot(userId),
  });
}

export function useDisable() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => disableUser(id),
    onSuccess: (_, userId) => {
      invalidateUserDetailCaches(queryClient, userId);
      toast.success("User disabled");
    },
  });
}

export function useEnable() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => enableUser(id),
    onSuccess: (_, userId) => {
      invalidateUserDetailCaches(queryClient, userId);
      toast.success("User enabled");
    },
  });
}

export function useResetPassword() {
  const queryClient = useQueryClient();
  const [tempPasswordData, setTempPasswordData] =
    useState<AdminResetPasswordResponse | null>(null);

  const mutation = useMutation({
    mutationFn: (id: string) => resetUserPassword(id),
    onSuccess: (data) => {
      setTempPasswordData(data);
      hydrateUserCaches(queryClient, data.user);
      invalidateUserDetailCaches(queryClient, data.user.id);
      toast.success("Temporary password generated");
    },
  });

  return {
    ...mutation,
    tempPasswordData,
    clearTempPasswordData: () => setTempPasswordData(null),
  };
}

export function useRevokeSessions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => revokeUserSessions(id),
    onSuccess: (_, userId) => {
      invalidateUserDetailCaches(queryClient, userId);
      toast.success("Sessions revoked");
    },
  });
}

export function useUserActivity(userId: string, input: UserActivityFilterInput) {
  const filter = userActivityFilterSchema.parse(input);
  return useQuery({
    queryKey: adminQueryKeys.userActivity(userId, filter),
    queryFn: () => listUserActivity(userId, filter),
    enabled: Boolean(userId),
  });
}
