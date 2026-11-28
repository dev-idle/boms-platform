import { Button } from "./button";

type DashboardLoadErrorProps = {
  message: string;
  /** Asks again; left out when asking again cannot help, e.g. the thing is gone. */
  onRetry?: () => void;
};

/** A detail page that could not load what it is about: what happened, and the way to ask again. */
export function DashboardLoadError({ message, onRetry }: DashboardLoadErrorProps) {
  return (
    <div className="grid justify-items-start gap-3">
      <p className="text-sm text-error">{message}</p>
      {onRetry ? (
        <Button size="sm" type="button" variant="outline" onClick={onRetry}>
          Try again
        </Button>
      ) : null}
    </div>
  );
}
