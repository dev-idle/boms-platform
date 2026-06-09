import type { ReactNode } from "react";

type AuthPageShellProps = {
  brandTitle: string;
  brandDescription: string;
  title: string;
  description: string;
  children: ReactNode;
  footer?: ReactNode;
};

export function AuthPageShell({
  brandTitle,
  brandDescription,
  title,
  description,
  children,
  footer,
}: AuthPageShellProps) {
  return (
    <div className="flex flex-1 flex-col lg:grid lg:grid-cols-2">
      <section
        aria-hidden="true"
        className="relative hidden overflow-hidden bg-surface-alt lg:flex lg:flex-col lg:justify-center"
      >
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 bg-linear-to-br from-primary-subtle via-surface-alt to-greige-100"
        />
        <div className="relative mx-auto max-w-md px-8 py-16 xl:px-12">
          <div className="h-px w-12 border-t border-accent" />
          <h2 className="mt-8 font-heading text-3xl font-medium leading-tight tracking-tight text-foreground xl:text-4xl">
            {brandTitle}
          </h2>
          <p className="mt-4 text-base leading-relaxed text-muted">
            {brandDescription}
          </p>
        </div>
      </section>

      <section className="flex flex-1 flex-col justify-center px-4 py-12 sm:px-6 lg:px-8">
        <div className="mx-auto w-full max-w-md">
          <div className="rounded-lg border border-border bg-surface p-8 sm:p-10">
            <div className="lg:hidden">
              <div className="h-px w-10 border-t border-accent" />
            </div>
            <h1 className="mt-4 font-heading text-2xl font-medium tracking-tight text-foreground lg:mt-0">
              {title}
            </h1>
            <p className="mt-2 text-sm leading-relaxed text-muted">
              {description}
            </p>
            <div className="mt-8">{children}</div>
          </div>
          {footer ? <div className="mt-6 text-center">{footer}</div> : null}
        </div>
      </section>
    </div>
  );
}
