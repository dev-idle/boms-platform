import { DotsRing } from "@/components/ui/dots-ring";
import { LOADING_MESSAGE } from "@/constants/loading-copy";

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

/**
 * No `className`: every wait of a kind is the same size, so none can drift
 * into a height of its own.
 */
function LoadingState({ variant }: { variant: "inline" | "page" }) {
  return (
    <div
      aria-busy="true"
      className={`loading-state loading-state--${variant}`}
      role="status"
    >
      <BusyIndicator />
    </div>
  );
}

/** Whole-page wait: centred in the viewport's free height. */
export function PageLoadingState() {
  return <LoadingState variant="page" />;
}

/** Wait inside a page or section: centred in one fixed-height block. */
export function InlineLoadingState() {
  return <LoadingState variant="inline" />;
}
