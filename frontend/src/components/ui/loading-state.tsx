import { DotsRing } from "@/components/ui/dots-ring";
import { LOADING_MESSAGE } from "@/constants/loading-copy";
import { cn } from "@/lib/utils";

/**
 * The one spinner: a dots ring (`@loading-ui/dots-ring`), always the same
 * number of dots, sized and coloured by `.loading-indicator` for the theme.
 * Decorative on its own; `BusyIndicator` adds the announcement.
 */
function LoadingIndicator() {
  return (
    <div aria-hidden="true" className="loading-indicator">
      <DotsRing className="loading-indicator__spinner" />
    </div>
  );
}

/** The spinner with its announcement: the unit every loading surface is built from. */
export function BusyIndicator() {
  return (
    <>
      <span className="sr-only">{LOADING_MESSAGE}</span>
      <LoadingIndicator />
    </>
  );
}

type LoadingStateProps = {
  className?: string;
};

function LoadingState({
  className,
  variant,
}: LoadingStateProps & { variant: "inline" | "page" }) {
  return (
    <div
      aria-busy="true"
      className={cn("loading-state", `loading-state--${variant}`, className)}
      role="status"
    >
      <BusyIndicator />
    </div>
  );
}

/** Whole-page wait: centred in the viewport's free height. */
export function PageLoadingState({ className }: LoadingStateProps) {
  return <LoadingState className={className} variant="page" />;
}

/** Wait inside a page or section: centred in a padded block. */
export function InlineLoadingState({ className }: LoadingStateProps) {
  return <LoadingState className={className} variant="inline" />;
}
