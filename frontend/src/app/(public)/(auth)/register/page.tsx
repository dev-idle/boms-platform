import { RegisterForm } from "@/features/auth";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";
import { validateNext } from "@/lib/validate-next";

export const metadata = pageTitle(PAGE_TITLES.createAccount);

type RegisterPageProps = {
  searchParams: Promise<{
    next?: string;
  }>;
};

export default async function RegisterPage({ searchParams }: RegisterPageProps) {
  const params = await searchParams;
  const safeNext = validateNext(params.next) ?? undefined;

  return <RegisterForm next={safeNext} />;
}
