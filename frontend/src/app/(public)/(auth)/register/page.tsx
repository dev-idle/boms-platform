import { Suspense } from "react";

import { AUTH_FORM_COPY, AuthFormSkeleton, RegisterForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.createAccount);

export default function RegisterPage() {
  return (
    <Suspense fallback={<AuthFormSkeleton {...AUTH_FORM_COPY.createAccount} />}>
      <RegisterForm />
    </Suspense>
  );
}
