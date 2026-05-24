"use client";

import { Button } from "@/components/ui/button";

import { useLogout } from "../hooks";

type LogoutButtonProps = {
  className?: string;
  variant?: "default" | "outline" | "ghost";
  size?: "default" | "sm" | "lg";
};

export function LogoutButton({
  className,
  variant = "ghost",
  size = "sm",
}: LogoutButtonProps) {
  const logout = useLogout();

  return (
    <Button
      className={className}
      disabled={logout.isPending}
      onClick={() => logout.mutate()}
      size={size}
      type="button"
      variant={variant}
    >
      {logout.isPending ? "Signing out…" : "Sign out"}
    </Button>
  );
}
