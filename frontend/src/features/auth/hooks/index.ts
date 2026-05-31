"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { validateNextForRole } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";
import { getMe as getMeFromUserApi } from "@/features/user/api";
import {
  homeRouteForRole,
  passwordRouteForRole,
} from "@/lib/routing/role-routes";

import { login, logout, register } from "../api";
import { resetRefreshManager, scheduleRefresh } from "@/lib/auth";
import type { LoginInput, RegisterInput } from "../schemas";
import { userQueryKeys } from "@/features/user/hooks/query-options";

export { meQueryOptions, userQueryKeys } from "@/features/user/hooks/query-options";
export { useMe } from "@/features/user/hooks";

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
  const clearAuth = useAuthStore((state) => state.clearAuth);

  return useMutation({
    mutationFn: async ({ input }: LoginMutationVariables) => {
      const data = await login(input);
      useAuthStore.getState().setTokens({
        accessToken: data.access_token,
        expiresIn: data.expires_in,
      });
      scheduleRefresh(useAuthStore.getState().expiresAt);

      try {
        const me = await getMeFromUserApi();
        return { data, me };
      } catch (error) {
        resetRefreshManager();
        clearAuth();
        throw error;
      }
    },
    onSuccess: ({ data, me }, variables) => {

      setAuth({
        accessToken: data.access_token,
        expiresIn: data.expires_in,
        user: me,
      });

      void queryClient.invalidateQueries({ queryKey: userQueryKeys.me });

      const mustChangePassword = data.must_change_password ?? me.must_change_password;
      if (mustChangePassword) {
        router.push(passwordRouteForRole(me.role));
        return;
      }

      router.push(
        validateNextForRole(variables.next, me.role) ?? homeRouteForRole(me.role),
      );
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
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      router.push(ROUTE.login);
    },
  });
}
