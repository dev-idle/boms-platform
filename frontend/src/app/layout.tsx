import type { Metadata, Viewport } from "next";
import { Suspense } from "react";
import { Toaster } from "sonner";

import "./globals.css";
import { PageLoadingState } from "@/components/ui/loading-state";
import { ThemeScope } from "@/components/theme/theme-scope";
import { BRAND } from "@/constants/brand";
import { APP_THEME } from "@/constants/themes";
import { AuthBootstrap } from "@/features/auth/server";
import { QueryProvider } from "@/providers";

import { display, instrument, mono } from "./fonts";

const appUrl =
  process.env.NEXT_PUBLIC_APP_URL?.replace(/\/$/, "") ?? "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(appUrl),
  title: {
    default: BRAND.name,
    template: `%s | ${BRAND.name}`,
  },
  description: BRAND.tagline,
  applicationName: BRAND.name,
  referrer: "strict-origin-when-cross-origin",
  openGraph: {
    title: BRAND.name,
    description: BRAND.tagline,
    siteName: BRAND.name,
    type: "website",
  },
  twitter: {
    card: "summary",
    title: BRAND.name,
    description: BRAND.tagline,
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  viewportFit: "cover",
  themeColor: "#FBF9F6",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      data-scroll-behavior="smooth"
      className={`${display.variable} ${instrument.variable} ${mono.variable} h-full`}
      suppressHydrationWarning
    >
      <body className="flex min-h-full flex-col bg-bg font-body text-ink-2">
        <QueryProvider>
          <Suspense
            fallback={
              <ThemeScope theme={APP_THEME.storefront}>
                <PageLoadingState />
              </ThemeScope>
            }
          >
            <AuthBootstrap>{children}</AuthBootstrap>
          </Suspense>
          <Toaster
            closeButton
            position="top-center"
            toastOptions={{
              classNames: {
                toast:
                  "border border-border bg-surface text-ink-2 shadow-overlay",
                title: "text-toast-title text-ink",
                description: "text-toast-description text-muted",
                success: "border-l-[3px] border-l-success",
                error: "border-l-[3px] border-l-error",
                warning: "border-l-[3px] border-l-warning",
                info: "border-l-[3px] border-l-info",
              },
            }}
          />
        </QueryProvider>
      </body>
    </html>
  );
}
