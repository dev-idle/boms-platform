"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

import {
  createOperational,
  disableUser,
  getUserById,
  listUsers,
  revokeUserSessions,
  updateRole,
  updateUserProfile,
} from "../api";
import {
  listFilterSchema,
  type AdminUser,
  type CreateOperationalInput,
  type CreateOperationalResponse,
  type ListFilterInput,
  type UpdateOperationalProfileInput,
  type UpdateRoleInput,
} from "../schemas";
import { adminQueryKeys } from "./query-options";

export { adminQueryKeys } from "./query-options";

function normalizeFilter(input: ListFilterInput): ListFilterInput {
  return listFilterSchema.parse(input);
}

function hydrateUserCaches(queryClient: ReturnType<typeof useQueryClient>, user: AdminUser): void {
  queryClient.setQueryData(adminQueryKeys.user(user.id), user);
}

export function useUsers(input: ListFilterInput) {
  const filter = normalizeFilter(input);
  return useQuery({
    queryKey: adminQueryKeys.users(filter),
    queryFn: () => listUsers(filter),
    placeholderData: keepPreviousData,
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

export function useUpdateUserProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: { id: string; input: UpdateOperationalProfileInput }) =>
      updateUserProfile(params.id, params.input),
    onSuccess: (user) => {
      hydrateUserCaches(queryClient, user);
      void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
      toast.success("Profile updated");
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: { id: string; input: UpdateRoleInput }) =>
      updateRole(params.id, params.input),
    onSuccess: (user) => {
      hydrateUserCaches(queryClient, user);
      void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
      toast.success("Role updated");
    },
  });
}

export function useDisable() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => disableUser(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
      toast.success("User disabled");
    },
  });
}

export function useRevokeSessions() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => revokeUserSessions(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: adminQueryKeys.usersRoot });
      toast.success("Sessions revoked");
    },
  });
}
