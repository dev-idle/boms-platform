/** Status glyphs for the app toaster: open strokes, no container, tinted by the toast. */

const TOAST_ICON_STROKE = 1.5;

type ToastIconProps = {
  className?: string;
};

export function ToastSuccessIcon({ className }: ToastIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={TOAST_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

export function ToastErrorIcon({ className }: ToastIconProps) {
  return (
    <svg
      aria-hidden
      className={className}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={TOAST_ICON_STROKE}
      viewBox="0 0 24 24"
    >
      <path d="M12 5v9" />
      <path d="M12 19h.01" />
    </svg>
  );
}
