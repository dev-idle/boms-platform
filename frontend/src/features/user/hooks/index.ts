"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { endLocalSession } from "@/lib/auth";
import { saveJsonFile } from "@/lib/dom/save-json-file";
import { isApiError } from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import { changePassword, deleteAccount, exportMyData, updateProfile } from "../api";
import { dataExportFileName } from "../lib/data-export-file-name";
import {
  type ChangePasswordInput,
  type EraseMyAccountInput,
  type UpdateSelfProfileInput,
} from "../schemas/index";
import { meQueryOptions, userQueryKeys } from "./query-options";

export { userQueryKeys } from "./query-options";

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
  const beginLogout = useAuthStore((state) => state.beginLogout);

  return useMutation({
    mutationFn: (input: ChangePasswordInput) => changePassword(input),
    // Signing out starts only once the password changed: a refused change (a
    // wrong current password) leaves the form on the page to say why.
    onSuccess: () => {
      beginLogout();
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      router.push(`${ROUTE.login}?changed=1`);
    },
  });
}

export function useDeleteAccount() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const beginLogout = useAuthStore((state) => state.beginLogout);

  return useMutation({
    mutationFn: (input: EraseMyAccountInput) => deleteAccount(input),
    // Signing out starts only once the account is gone: a refused erasure leaves
    // the customer on the page, where the form says why.
    onSuccess: () => {
      beginLogout();
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      toast.success("Your account and personal details were erased");
      router.push(ROUTE.login);
    },
  });
}

/** Downloads everything the bakery holds about the signed-in person as a JSON file. */
export function useDownloadMyData() {
  return useMutation({
    mutationFn: () => exportMyData(),
    onSuccess: (data) => {
      saveJsonFile(dataExportFileName(new Date(data.exported_at)), data);
      toast.success("Your data is downloading");
    },
    onError: (error) => {
      toast.error(
        isApiError(error) && error.isRateLimited()
          ? "You have downloaded your data several times recently. Please try again later."
          : "We could not prepare your data. Please try again.",
      );
    },
  });
}
