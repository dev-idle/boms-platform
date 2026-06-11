import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

type IconProps = {
  className?: string;
};

const iconProps = {
  fill: "none",
  stroke: "currentColor",
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
  strokeWidth: 1.75,
  viewBox: "0 0 24 24",
  xmlns: "http://www.w3.org/2000/svg",
};

function IconBase({
  children,
  className,
}: IconProps & { children: ReactNode }) {
  return (
    <svg aria-hidden className={cn("dashboard-nav-icon-svg", className)} {...iconProps}>
      {children}
    </svg>
  );
}

export function DashboardIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <rect height="7" rx="1" width="7" x="3" y="3" />
      <rect height="7" rx="1" width="7" x="14" y="3" />
      <rect height="7" rx="1" width="7" x="14" y="14" />
      <rect height="7" rx="1" width="7" x="3" y="14" />
    </IconBase>
  );
}

export function UsersIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
    </IconBase>
  );
}

export function OrdersIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M9 5H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-4" />
      <path d="M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v0a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2z" />
      <path d="M9 12h6M9 16h6" />
    </IconBase>
  );
}

export function CategoriesIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M4 7h16M4 12h10M4 17h14" />
    </IconBase>
  );
}

export function ProductsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
      <path d="M3.3 7.7 12 12.5l8.7-4.8M12 22V12.5" />
    </IconBase>
  );
}

export function CombosIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="m7.5 4.27 9 5.15M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16V8z" />
      <path d="M3.3 7.7 12 12.5l8.7-4.8M12 22V12.5" />
    </IconBase>
  );
}

export function DiscountsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
      <path d="M7 7h.01" />
    </IconBase>
  );
}

export function ProfileIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </IconBase>
  );
}

export function PasswordIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <rect height="11" rx="2" width="18" x="3" y="11" />
      <path d="M7 11V7a5 5 0 0 1 10 0v4" />
    </IconBase>
  );
}

export function LogOutIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
      <path d="M16 17l5-5-5-5M21 12H9" />
    </IconBase>
  );
}

export const DASHBOARD_NAV_ICONS = {
  categories: CategoriesIcon,
  combos: CombosIcon,
  dashboard: DashboardIcon,
  discounts: DiscountsIcon,
  orders: OrdersIcon,
  password: PasswordIcon,
  products: ProductsIcon,
  profile: ProfileIcon,
  users: UsersIcon,
} as const;

export type DashboardNavIconId = keyof typeof DASHBOARD_NAV_ICONS;

export function DashboardNavIcon({
  className,
  icon,
}: {
  className?: string;
  icon: DashboardNavIconId;
}) {
  const Icon = DASHBOARD_NAV_ICONS[icon];
  return <Icon className={className} />;
}
