"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { validateNext } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { login, logout, register } from "../api";
import { resetRefreshManager, scheduleRefresh } from "@/lib/auth";
import type { LoginInput, RegisterInput } from "../schemas";
import { authQueryKeys, meQueryOptions } from "./query-options";

export { authQueryKeys, meQueryOptions } from "./query-options";

export function useMe() {
  const status = useAuthStore((state) => state.status);
  return useQuery({
    ...meQueryOptions(),
    enabled: status === "authenticated",
  });
}

export function useRegister() {
  const router = useRouter();

  return useMutation({
    mutationFn: (input: RegisterInput) => register(input),
    onSuccess: () => {
      router.push(`${ROUTE.login}?registered=1`);
    },
    onError: (error) => {
      if (!isApiError(error)) {
        toast.error("Something went wrong. Please try again.");
      }
    },
  });
}

export type LoginMutationVariables = {
  input: LoginInput;
  next?: string | null;
};

export function useLogin() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const setAuth = useAuthStore((state) => state.setAuth);

  return useMutation({
    mutationFn: ({ input }: LoginMutationVariables) => login(input),
    onSuccess: (data, variables) => {
      setAuth({
        accessToken: data.access_token,
        expiresIn: data.expires_in,
        user: data.user,
      });
      scheduleRefresh(useAuthStore.getState().expiresAt);
      void queryClient.invalidateQueries({ queryKey: authQueryKeys.me });
      router.push(validateNext(variables.next) ?? ROUTE.home);
    },
  });
}

export function useLogout() {
  const router = useRouter();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => logout(),
    onSettled: () => {
      resetRefreshManager();
      clearAuth();
      queryClient.removeQueries({ queryKey: authQueryKeys.me });
      router.push(ROUTE.login);
    },
  });
}
