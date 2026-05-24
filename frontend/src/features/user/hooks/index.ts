"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { resetRefreshManager } from "@/lib/auth";
import { useAuthStore } from "@/stores/auth-store";

import { changePassword, deleteAccount, updateProfile } from "../api";
import { type ChangePasswordInput, type UpdateSelfProfileInput } from "../schemas/index";
import type { Me } from "../types";
import { meQueryOptions, userQueryKeys } from "./query-options";

export { meQueryOptions, userQueryKeys } from "./query-options";

function applySelfProfilePatch(
  current: Me,
  input: UpdateSelfProfileInput,
): Me {
  switch (current.role) {
    case "customer":
      return {
        ...current,
        profile: {
          ...current.profile,
          display_name:
            input.display_name === undefined
              ? current.profile.display_name
              : input.display_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    case "staff":
    case "baker":
    case "manager":
      return {
        ...current,
        profile: {
          ...current.profile,
          full_name:
            input.full_name === undefined
              ? current.profile.full_name
              : input.full_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    case "admin":
      return {
        ...current,
        profile: {
          ...current.profile,
          full_name:
            input.full_name === undefined
              ? current.profile.full_name
              : input.full_name,
          phone:
            input.phone === undefined ? current.profile.phone : input.phone,
        },
      };
    default: {
      throw new Error(`Unhandled user role payload: ${String(current)}`);
    }
  }
}

export function useMe() {
  const status = useAuthStore((state) => state.status);
  const updateUser = useAuthStore((state) => state.updateUser);

  const query = useQuery({
    ...meQueryOptions(),
    enabled: status === "authenticated",
  });

  useEffect(() => {
    if (query.data) {
      updateUser(query.data);
    }
  }, [query.data, updateUser]);

  return query;
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  const updateUserStore = useAuthStore((state) => state.updateUser);

  return useMutation({
    mutationFn: (input: UpdateSelfProfileInput) => updateProfile(input),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: userQueryKeys.me });

      const previousUser = useAuthStore.getState().user;
      if (previousUser) {
        const optimistic = applySelfProfilePatch(previousUser, input);
        updateUserStore(optimistic);
        queryClient.setQueryData(userQueryKeys.me, optimistic);
      }

      return { previousUser };
    },
    onError: (_error, _input, context) => {
      if (context?.previousUser) {
        updateUserStore(context.previousUser);
        queryClient.setQueryData(userQueryKeys.me, context.previousUser);
      }
    },
    onSuccess: (me) => {
      updateUserStore(me);
      queryClient.setQueryData(userQueryKeys.me, me);
      toast.success("Profile updated");
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: userQueryKeys.me });
    },
  });
}

export function useChangePassword() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const clearAuth = useAuthStore((state) => state.clearAuth);

  return useMutation({
    mutationFn: (input: ChangePasswordInput) => changePassword(input),
    onSuccess: () => {
      resetRefreshManager();
      clearAuth();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      toast.success("Password changed. Please sign in again.");
      router.push(`${ROUTE.login}?changed=1`);
    },
  });
}

export function useDeleteAccount() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const clearAuth = useAuthStore((state) => state.clearAuth);

  return useMutation({
    mutationFn: () => deleteAccount(),
    onSuccess: () => {
      resetRefreshManager();
      clearAuth();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      toast.success("Account deleted");
      router.push(ROUTE.login);
    },
  });
}
