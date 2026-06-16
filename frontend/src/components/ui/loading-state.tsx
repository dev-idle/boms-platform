import { DotsRing } from "@/components/ui/dots-ring";
import { LOADING_MESSAGE } from "@/constants/loading-copy";
import { cn } from "@/lib/utils";

type LoadingStateVariant = "compact" | "inline" | "page";

type LoadingIndicatorProps = {
  className?: string;
  /** Fewer dots for dense surfaces (e.g. table cells). */
  dots?: number;
};

/** Dots ring spinner — `@loading-ui/dots-ring` via shadcn. */
export function LoadingIndicator({ className, dots }: LoadingIndicatorProps) {
  return (
    <div aria-hidden="true" className={cn("loading-indicator", className)}>
      <DotsRing className="loading-indicator__spinner" dots={dots} />
    </div>
  );
}

type LoadingStateProps = {
  className?: string;
  message?: string;
  showMessage?: boolean;
  variant?: LoadingStateVariant;
};

/** Centered loading — dots ring; copy hidden by default (`aria-label` still set). */
export function LoadingState({
  className,
  message = LOADING_MESSAGE,
  showMessage = false,
  variant = "page",
}: LoadingStateProps) {
  return (
    <div
      aria-busy="true"
      aria-label={message}
      aria-live="polite"
      className={cn("loading-state", `loading-state--${variant}`, className)}
      role="status"
    >
      <LoadingIndicator />
      {showMessage ? (
        <p className="loading-state__message">{message}</p>
      ) : null}
    </div>
  );
}

type PageLoadingStateProps = {
  className?: string;
  message?: string;
  showMessage?: boolean;
};

export function PageLoadingState(props: PageLoadingStateProps) {
  return <LoadingState {...props} variant="page" />;
}

type InlineLoadingStateProps = {
  className?: string;
  message?: string;
  showMessage?: boolean;
  variant?: Extract<LoadingStateVariant, "compact" | "inline">;
};

export function InlineLoadingState({
  variant = "inline",
  ...props
}: InlineLoadingStateProps) {
  return <LoadingState {...props} variant={variant} />;
}
