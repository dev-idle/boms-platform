import { cookies } from "next/headers";
import type { ReactNode } from "react";

import { AUTH_REFRESH_COOKIE } from "@/constants/cookies";

import { AuthProvider } from "./auth-provider";

type AuthBootstrapProps = {
  children: ReactNode;
};

export async function AuthBootstrap({ children }: AuthBootstrapProps) {
  const cookieStore = await cookies();
  const initialAuthHint = cookieStore.has(AUTH_REFRESH_COOKIE);

  return (
    <AuthProvider initialAuthHint={initialAuthHint}>{children}</AuthProvider>
  );
}
