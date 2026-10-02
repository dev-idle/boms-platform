import { cn } from "@/lib/utils";

type IconProps = {
  className?: string;
};

const STOREFRONT_ICON_VIEWBOX = "0 0 24 24";

export function SearchIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20 16.5 16.5" />
    </svg>
  );
}

export function CartIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="18"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.4"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="18"
    >
      <path d="M6.5 8h11l1.2 11.2a1.6 1.6 0 0 1-1.6 1.8H6.9a1.6 1.6 0 0 1-1.6-1.8L6.5 8Z" />
      <path d="M9.2 8V6.6a2.8 2.8 0 0 1 5.6 0V8" />
    </svg>
  );
}

export function OrdersIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <path d="M8 3h8l1.25 4.25V19a2 2 0 0 1-2 2H8.75a2 2 0 0 1-2-2V7.25L8 3z" />
      <path d="M8 7.5h8" />
      <path d="M9.5 11.5h5" />
      <path d="M9.5 15h5" />
    </svg>
  );
}

export function CloseIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[1.125rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <path d="M6 6l12 12" />
      <path d="M18 6 6 18" />
    </svg>
  );
}

export function ChevronDownIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[0.875rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}

export function UserIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <circle cx="12" cy="8.25" r="3.25" />
      <path d="M5.25 20.25c0-3.35 2.95-5.75 6.75-5.75s6.75 2.4 6.75 5.75" />
    </svg>
  );
}
export function SignOutIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
      <path d="m16 17 5-5-5-5" />
      <path d="M21 12H9" />
    </svg>
  );
}

/** A product kept among favorites; filled while it is on the list. */
export function HeartIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <path d="M12 20s-7.5-4.6-7.5-10.1A4.15 4.15 0 0 1 12 7.4a4.15 4.15 0 0 1 7.5 2.5C19.5 15.4 12 20 12 20Z" />
    </svg>
  );
}

/** A product kept on the wishlist; filled while it is on the list. */
export function BookmarkIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <path d="M6.5 4h11v16.5L12 16.75 6.5 20.5V4Z" />
    </svg>
  );
}

/** One star of a rating; filled by the rating's own styles. */
export function StarIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("storefront-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox={STOREFRONT_ICON_VIEWBOX}
      width="24"
    >
      <path d="M12 3.75 14.55 8.9l5.7.83-4.13 4.02.98 5.67L12 16.74l-5.1 2.68.98-5.67-4.13-4.02 5.7-.83L12 3.75Z" />
    </svg>
  );
}
