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

function DashboardIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <rect height="7" rx="1" width="7" x="3" y="3" />
      <rect height="7" rx="1" width="7" x="14" y="3" />
      <rect height="7" rx="1" width="7" x="14" y="14" />
      <rect height="7" rx="1" width="7" x="3" y="14" />
    </IconBase>
  );
}

function UsersIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
    </IconBase>
  );
}

function OrdersIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M9 5H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-4" />
      <path d="M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v0a2 2 0 0 1-2 2h-2a2 2 0 0 1-2-2z" />
      <path d="M9 12h6M9 16h6" />
    </IconBase>
  );
}

function CategoriesIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M4 7h16M4 12h10M4 17h14" />
    </IconBase>
  );
}

function ProductsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
      <path d="M3.3 7.7 12 12.5l8.7-4.8M12 22V12.5" />
    </IconBase>
  );
}

function CombosIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="m7.5 4.27 9 5.15M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16V8z" />
      <path d="M3.3 7.7 12 12.5l8.7-4.8M12 22V12.5" />
    </IconBase>
  );
}

function DiscountsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
      <path d="M7 7h.01" />
    </IconBase>
  );
}

/** Checklist — a queue of tickets to work through. */
function PrepIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M10 6h10M10 12h10M10 18h10" />
      <path d="m4 6 1.5 1.5L8 5M4 12l1.5 1.5L8 11M4 18l1.5 1.5L8 17" />
    </IconBase>
  );
}

/** A tray with a slash — what has run out today. */
function AvailabilityIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M3 14h18v5H3z" />
      <path d="M7 14V9h10v5" />
      <path d="M4 4l16 16" />
    </IconBase>
  );
}

/** A speech bubble — customers writing about their orders. */
function MessagesIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M21 12a8 8 0 0 1-8 8H7l-4 3V12a8 8 0 0 1 8-8h2a8 8 0 0 1 8 8Z" />
    </IconBase>
  );
}

/** Envelope — promotions emailed to customers. */
function PromotionsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <rect height="14" width="18" x="3" y="5" />
      <path d="m3 7 9 6 9-6" />
    </IconBase>
  );
}

/** Heart — how customers take to the shop's features. */
function EngagementIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M12 20s-7.5-4.6-7.5-10.1A4.15 4.15 0 0 1 12 7.4a4.15 4.15 0 0 1 7.5 2.5C19.5 15.4 12 20 12 20Z" />
    </IconBase>
  );
}

/** Warning triangle — what went wrong with orders. */
function IncidentsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M12 3.5 21.5 20h-19L12 3.5Z" />
      <path d="M12 10v4.5M12 17.25v.01" />
    </IconBase>
  );
}

/** Star — what customers rate their pickups. */
function ReviewsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M12 3.75 14.55 8.9l5.7.83-4.13 4.02.98 5.67L12 16.74l-5.1 2.68.98-5.67-4.13-4.02 5.7-.83L12 3.75Z" />
    </IconBase>
  );
}

/** Clock — the day's pickup times. */
function PickupsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </IconBase>
  );
}

function ProfileIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </IconBase>
  );
}

/** Sliders — adjustable settings. */
function SettingsIcon({ className }: IconProps) {
  return (
    <IconBase className={className}>
      <path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h8M16 18h4" />
      <circle cx="16" cy="6" r="2" />
      <circle cx="10" cy="12" r="2" />
      <circle cx="14" cy="18" r="2" />
    </IconBase>
  );
}

function PasswordIcon({ className }: IconProps) {
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

const DASHBOARD_NAV_ICONS = {
  availability: AvailabilityIcon,
  categories: CategoriesIcon,
  combos: CombosIcon,
  dashboard: DashboardIcon,
  discounts: DiscountsIcon,
  engagement: EngagementIcon,
  incidents: IncidentsIcon,
  messages: MessagesIcon,
  orders: OrdersIcon,
  password: PasswordIcon,
  pickups: PickupsIcon,
  prep: PrepIcon,
  products: ProductsIcon,
  profile: ProfileIcon,
  promotions: PromotionsIcon,
  reviews: ReviewsIcon,
  settings: SettingsIcon,
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
