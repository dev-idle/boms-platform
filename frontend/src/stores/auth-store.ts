import { create } from "zustand";

type AuthState = {
  /** Client-visible auth phase only — never store tokens in this store. */
  status: "unknown" | "signedOut" | "signedIn";
  setStatus: (status: AuthState["status"]) => void;
};

export const useAuthUiStore = create<AuthState>((set) => ({
  status: "unknown",
  setStatus: (status) => set({ status }),
}));
