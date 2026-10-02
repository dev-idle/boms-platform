"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { loginHrefAfterRegister } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";
import {
  getMe as getMeFromUserApi,
  meQueryOptions,
  primeMeQueryCache,
  userQueryKeys,
} from "@/features/user";
import { resolvePostAuthDestination } from "@/lib/routing/post-auth-destination";

import {
  confirmPasswordReset,
  login,
  logout,
  register,
  requestPasswordReset,
  unsubscribe,
  verifyEmail,
} from "../api";
import { endLocalSession, scheduleRefresh } from "@/lib/auth";
import type { ForgotPasswordInput, LoginInput, PasswordResetConfirmInput, RegisterInput } from "../schemas";

export { useAuthHydrated } from "./use-auth-hydrated";
export { useLinkToken } from "./use-link-token";

export function useRequestPasswordReset() {
  return useMutation({
    mutationFn: (input: ForgotPasswordInput) => requestPasswordReset(input),
  });
}

/**
 * The reset ends every session of the account, this browser's too, so a
 * signed-in holder is signed out here and sent to sign in with the new password.
 */
export function useConfirmPasswordReset() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: PasswordResetConfirmInput) => confirmPasswordReset(input),
    onSuccess: () => {
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      router.replace(`${ROUTE.login}?reset=1`);
    },
  });
}

/** Stops promotion emails from a link; a signed-in customer's account shows it at once. */
export function useUnsubscribe() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (token: string) => unsubscribe(token),
    onSuccess: () => {
      if (useAuthStore.getState().status === "authenticated") {
        void queryClient.prefetchQuery({ ...meQueryOptions(), staleTime: 0 });
      }
    },
  });
}

/**
 * Confirms the address. The link may be another account's, so a signed-in
 * holder's /me is read again in the background; the result does not wait for
 * it, and the pages that depend on it read it again themselves.
 */
export function useVerifyEmail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (token: string) => verifyEmail(token),
    onSuccess: () => {
      if (useAuthStore.getState().status === "authenticated") {
        void queryClient.prefetchQuery({ ...meQueryOptions(), staleTime: 0 });
      }
    },
  });
}

export type RegisterMutationVariables = {
  input: RegisterInput;
  next?: string | null;
};

export function useRegister() {
  const router = useRouter();

  return useMutation({
    mutationFn: ({ input }: RegisterMutationVariables) => register(input),
    onSuccess: (_data, variables) => {
      router.push(loginHrefAfterRegister(variables.next));
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
        endLocalSession();
        throw error;
      }
    },
    onSuccess: ({ data, me }, variables) => {
      setAuth({
        accessToken: data.access_token,
        expiresIn: data.expires_in,
        user: me,
      });

      primeMeQueryCache(queryClient, me);

      const mustChangePassword =
        data.must_change_password ?? me.must_change_password;

      router.replace(
        resolvePostAuthDestination(me.role, {
          next: variables.next,
          mustChangePassword,
        }),
      );
    },
  });
}

export function useLogout() {
  const router = useRouter();
  const beginLogout = useAuthStore((state) => state.beginLogout);
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => logout(),
    onMutate: () => {
      beginLogout();
    },
    onSettled: () => {
      endLocalSession();
      queryClient.removeQueries({ queryKey: userQueryKeys.me });
      router.replace(ROUTE.login);
    },
  });
}
