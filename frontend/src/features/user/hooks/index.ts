"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { endLocalSession } from "@/lib/auth";
import { useAuthStore } from "@/stores/auth-store";

import { changePassword, deleteAccount, updateProfile } from "../api";
import { type ChangePasswordInput, type UpdateSelfProfileInput } from "../schemas/index";
import { meQueryOptions, userQueryKeys } from "./query-options";

export { meQueryOptions, userQueryKeys } from "./query-options";

export function useMe() {
  const status = useAuthStore((state) => state.status);
  const updateUser = useAuthStore((state) => state.updateUser);

  const query = useQuery({
    ...meQueryOptions(),
    enabled: status === "authenticated",
    // Bootstrap/login already fetch /me into the auth store; skip a redundant mount fetch.
    refetchOnMount: () => useAuthStore.getState().user === null,
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
    onSuccess: (me) => {
      updateUserStore(me);
      queryClient.setQueryData(userQueryKeys.me, me);
      toast.success("Profile updated");
    },
  });
}

export function useChangePassword() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ChangePasswordInput) => changePassword(input),
    onSuccess: () => {
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      toast.success("Password changed. Please sign in again.");
      router.push(`${ROUTE.login}?changed=1`);
    },
  });
}

export function useDeleteAccount() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteAccount(),
    onSuccess: () => {
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      toast.success("Account deleted");
      router.push(ROUTE.login);
    },
  });
}
