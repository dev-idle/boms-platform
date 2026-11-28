import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";

type StorefrontEmptyStateProps = {
  title: string;
  message?: string;
  /** What to do from here, such as a link back to the shop. */
  action?: ReactNode;
};

/** Nothing to show yet: one framed message, and the way on (02-COMPONENTS §12). */
export function StorefrontEmptyState({ title, message, action }: StorefrontEmptyStateProps) {
  return (
    <div className="storefront-state">
      <p className="storefront-state__title">{title}</p>
      {message ? <p className="storefront-state__message">{message}</p> : null}
      {action ? <div className="storefront-state__action">{action}</div> : null}
    </div>
  );
}

type StorefrontErrorStateProps = {
  title: string;
  message?: string;
  onRetry: () => void;
};

/** A read that failed: what went wrong, and asking again. */
export function StorefrontErrorState({ title, message, onRetry }: StorefrontErrorStateProps) {
  return (
    <div className="storefront-state storefront-state--error" role="alert">
      <p className="storefront-state__title">{title}</p>
      {message ? <p className="storefront-state__message">{message}</p> : null}
      <Button onClick={onRetry} size="sm" type="button" variant="outline">
        Try again
      </Button>
    </div>
  );
}
