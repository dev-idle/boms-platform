"use client";

import { createContext, useContext, type ReactNode } from "react";

const SessionHintContext = createContext(false);

type SessionHintProviderProps = {
  hasRefreshCookie: boolean;
  children: ReactNode;
};

export function SessionHintProvider({
  hasRefreshCookie,
  children,
}: SessionHintProviderProps) {
  return (
    <SessionHintContext.Provider value={hasRefreshCookie}>
      {children}
    </SessionHintContext.Provider>
  );
}

/** Server-read refresh cookie presence (from AuthBootstrap — single cookies() call). */
export function useSessionAuthHint(): boolean {
  return useContext(SessionHintContext);
}
