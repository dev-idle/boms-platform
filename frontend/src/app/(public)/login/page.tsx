import { LoginForm } from "@/features/auth";
import { validateNext } from "@/lib/validate-next";

type LoginPageProps = {
  searchParams: Promise<{
    next?: string;
    registered?: string;
  }>;
};

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const safeNext = validateNext(params.next) ?? undefined;

  return (
    <LoginForm
      next={safeNext}
      registered={params.registered === "1"}
    />
  );
}
