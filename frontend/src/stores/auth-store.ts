import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

import type { UserRole } from "@/constants/roles";
import type { Me } from "@/features/user";

export type AuthStatus = "idle" | "authenticated" | "unauthenticated";

export type UserSummary = {
  email: string;
  role: UserRole;
};

type AuthState = {
  user: Me | null;
  userSummary: UserSummary | null;
  accessToken: string | null;
  expiresAt: number | null;
  status: AuthStatus;
  /** True while Sign out runs — RoleGate must not add ?next= on redirect. */
  logoutIntent: boolean;
  setAuth: (params: {
    accessToken: string;
    expiresIn: number;
    user: Me;
  }) => void;
  setTokens: (params: { accessToken: string; expiresIn: number }) => void;
  clearAuth: () => void;
  beginLogout: () => void;
  clearLogoutIntent: () => void;
  updateUser: (user: Me) => void;
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
      logoutIntent: false,
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
      beginLogout: () => set({ logoutIntent: true }),
      clearLogoutIntent: () => set({ logoutIntent: false }),
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

/** Wait until sessionStorage rehydration finishes (avoids false logout on F5). */
export function waitForAuthStoreHydration(): Promise<void> {
  if (useAuthStore.persist.hasHydrated()) {
    return Promise.resolve();
  }
  return new Promise((resolve) => {
    const unsub = useAuthStore.persist.onFinishHydration(() => {
      unsub();
      resolve();
    });
  });
}
