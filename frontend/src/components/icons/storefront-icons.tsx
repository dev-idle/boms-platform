import { cn } from "@/lib/utils";

type IconProps = {
  className?: string;
};

export function SearchIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[1.375rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="M20 20 16.5 16.5" />
    </svg>
  );
}

export function HeartIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[1.375rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <path d="M12 20.5s-6.5-4.35-8.5-8.25C1.9 8.9 3.4 5.5 6.6 5c1.55-.25 3.05.45 4.05 1.75.95-1.3 2.45-2 4-1.75 3.2.5 4.7 3.9 3.1 7.25-2 3.9-8.5 8.25-8.5 8.25z" />
    </svg>
  );
}

export function CartIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[1.375rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <path d="M6 3 3.5 6.5v12.75A1.75 1.75 0 0 0 5.25 21h13.5A1.75 1.75 0 0 0 20.5 19.25V6.5L18 3" />
      <path d="M3.5 6.5h17" />
      <path d="M15.75 6.5a3.75 3.75 0 0 1-7.5 0" />
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

export function UserIcon({ className }: IconProps) {
  return (
    <svg
      aria-hidden="true"
      className={cn("size-[1.375rem]", className)}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
    >
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5 20c0-3.5 3-6 7-6s7 2.5 7 6" />
    </svg>
  );
}
