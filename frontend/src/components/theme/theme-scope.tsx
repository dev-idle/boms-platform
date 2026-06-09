import type { ReactNode } from "react";

import type { AppTheme } from "@/constants/themes";
import { cn } from "@/lib/utils";

type ThemeScopeProps = {
  theme: AppTheme;
  children: ReactNode;
  className?: string;
};

/** Route-group root: sets `data-theme` for scoped semantic tokens in `globals.css`. */
export function ThemeScope({ theme, children, className }: ThemeScopeProps) {
  return (
    <div
      data-theme={theme}
      className={cn("min-h-full bg-bg text-foreground", className)}
    >
      {children}
    </div>
  );
}
