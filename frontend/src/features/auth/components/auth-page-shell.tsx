import type { ReactNode } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";

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
        className="relative hidden bg-blush lg:flex lg:flex-col lg:justify-center"
      >
        <div className="relative mx-auto max-w-md px-8 py-16 xl:px-12">
          <BrandLogo linked={false} size="lg" />
          <h2 className="mt-8 font-heading text-3xl font-medium leading-tight tracking-tight text-ink xl:text-4xl">
            {brandTitle}
          </h2>
          <p className="mt-4 text-base leading-relaxed text-ink-2">
            {brandDescription}
          </p>
        </div>
      </section>

      <section className="flex flex-1 flex-col justify-center bg-bg px-4 py-12 sm:px-6 lg:px-8">
        <div className="mx-auto w-full max-w-md">
          <div className="rounded-card bg-surface p-8 shadow-rest sm:p-10">
            <div className="lg:hidden">
              <BrandLogo linked={false} size="md" />
            </div>
            <h1 className="mt-6 font-heading text-2xl font-medium tracking-tight text-ink lg:mt-0">
              {title}
            </h1>
            <p className="mt-2 text-sm leading-relaxed text-ink-2">
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
