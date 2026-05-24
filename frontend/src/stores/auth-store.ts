import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

import type { UserRole } from "@/constants/roles";
import type { User } from "@/features/auth/schemas";

export type AuthStatus = "idle" | "authenticated" | "unauthenticated";

export type UserSummary = {
  email: string;
  role: UserRole;
};

type AuthState = {
  user: User | null;
  userSummary: UserSummary | null;
  accessToken: string | null;
  expiresAt: number | null;
  status: AuthStatus;
  setAuth: (params: {
    accessToken: string;
    expiresIn: number;
    user: User;
  }) => void;
  setTokens: (params: { accessToken: string; expiresIn: number }) => void;
  clearAuth: () => void;
  updateUser: (user: User) => void;
  setStatus: (status: AuthStatus) => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      userSummary: null,
      accessToken: null,
      expiresAt: null,
      status: "idle",
      setAuth: ({ accessToken, expiresIn, user }) =>
        set({
          accessToken,
          expiresAt: Date.now() + expiresIn * 1000,
          user,
          userSummary: { email: user.email, role: user.role },
          status: "authenticated",
        }),
      setTokens: ({ accessToken, expiresIn }) =>
        set({
          accessToken,
          expiresAt: Date.now() + expiresIn * 1000,
        }),
      clearAuth: () =>
        set({
          user: null,
          userSummary: null,
          accessToken: null,
          expiresAt: null,
          status: "unauthenticated",
        }),
      updateUser: (user) =>
        set({
          user,
          userSummary: { email: user.email, role: user.role },
        }),
      setStatus: (status) => set({ status }),
    }),
    {
      name: "boms-auth-summary",
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        userSummary: state.userSummary,
      }),
    },
  ),
);
