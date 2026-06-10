import type { ReactNode } from "react";
import Link from "next/link";

import { cn } from "@/lib/utils";

type AuthInlineLinkProps = {
  children: ReactNode;
  className?: string;
  href: string;
};

export function AuthInlineLink({
  children,
  className,
  href,
}: AuthInlineLinkProps) {
  return (
    <Link className={cn("auth-inline-link", className)} href={href}>
      {children}
    </Link>
  );
}
