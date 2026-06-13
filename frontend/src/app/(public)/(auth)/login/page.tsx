import { LoginForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";
import { validateNext } from "@/lib/validate-next";

export const metadata = pageTitle(PAGE_TITLES.signIn);

type LoginPageProps = {
  searchParams: Promise<{
    next?: string;
    registered?: string;
    changed?: string;
  }>;
};

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const safeNext = validateNext(params.next) ?? undefined;

  return (
    <LoginForm
      next={safeNext}
      registered={params.registered === "1"}
      changed={params.changed === "1"}
    />
  );
}
