import { Suspense } from "react";

import { AUTH_FORM_COPY, AuthFormSkeleton, LoginForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.signIn);

export default function LoginPage() {
  return (
    <Suspense fallback={<AuthFormSkeleton {...AUTH_FORM_COPY.signIn} />}>
      <LoginForm />
    </Suspense>
  );
}
