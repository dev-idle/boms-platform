import type { Metadata, Viewport } from "next";
import { Suspense, type CSSProperties } from "react";
import { Toaster } from "sonner";

import "./globals.css";
import { ToastErrorIcon, ToastSuccessIcon } from "@/components/icons/toast-icons";
import { RootRouteLoading } from "@/components/ui/route-loading";
import { BRAND } from "@/constants/brand";
import { AuthBootstrap } from "@/features/auth/server";
import { QueryProvider } from "@/providers";

import { display, instrument, mono } from "./fonts";

/** One value drives both sonner's timer and the drain on the toast's bottom hairline. */
const TOAST_DURATION_MS = 3500;
const TOASTER_STYLE = {
  "--app-toast-duration": `${TOAST_DURATION_MS}ms`,
} as CSSProperties;

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
          <Suspense fallback={<RootRouteLoading />}>
            <AuthBootstrap>{children}</AuthBootstrap>
          </Suspense>
          <Toaster
            closeButton
            duration={TOAST_DURATION_MS}
            icons={{ error: <ToastErrorIcon />, success: <ToastSuccessIcon /> }}
            position="bottom-right"
            style={TOASTER_STYLE}
          />
        </QueryProvider>
      </body>
    </html>
  );
}
