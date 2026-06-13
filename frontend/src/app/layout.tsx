import type { Metadata, Viewport } from "next";
import { Suspense } from "react";
import { Toaster } from "sonner";

import "./globals.css";
import { BRAND } from "@/constants/brand";
import { AuthBootstrap } from "@/features/auth/server";
import { QueryProvider } from "@/providers";

import { fraunces, instrument, mono } from "./fonts";

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
  themeColor: "#FBFAF9",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${fraunces.variable} ${instrument.variable} ${mono.variable} h-full`}
      suppressHydrationWarning
    >
      <body className="flex min-h-full flex-col bg-bg font-body text-ink-2">
        <QueryProvider>
          <Suspense
            fallback={
              <div className="flex flex-1 items-center justify-center text-caption">
                Loading…
              </div>
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
                  "rounded-card border border-border bg-surface text-ink-2 shadow-rest",
                title: "text-toast-title text-ink",
                description: "text-toast-description text-muted",
                success: "border-l-[3px] border-l-success !rounded-l-none",
                error: "border-l-[3px] border-l-error !rounded-l-none",
                warning: "border-l-[3px] border-l-warning !rounded-l-none",
                info: "border-l-[3px] border-l-info !rounded-l-none",
              },
            }}
          />
        </QueryProvider>
      </body>
    </html>
  );
}
