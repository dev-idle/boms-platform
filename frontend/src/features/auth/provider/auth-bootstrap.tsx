import { cookies } from "next/headers";
import { connection } from "next/server";
import type { ReactNode } from "react";

import { AUTH_REFRESH_COOKIE } from "@/constants/cookies";

import { AuthProvider } from "./auth-provider";

type AuthBootstrapProps = {
  children: ReactNode;
};

export async function AuthBootstrap({ children }: AuthBootstrapProps) {
  await connection();
  const cookieStore = await cookies();
  const hasRefreshCookie = cookieStore.has(AUTH_REFRESH_COOKIE);

  return (
    <AuthProvider hasRefreshCookie={hasRefreshCookie}>{children}</AuthProvider>
  );
}
