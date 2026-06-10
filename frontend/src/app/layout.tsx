import type { Metadata, Viewport } from "next";
import { Fraunces, Inter } from "next/font/google";
import { Suspense } from "react";
import { Toaster } from "sonner";

import "./globals.css";
import { BRAND } from "@/constants/brand";
import { AuthBootstrap } from "@/features/auth/server";
import { QueryProvider } from "@/providers";

const fraunces = Fraunces({
  variable: "--font-fraunces",
  subsets: ["latin"],
  axes: ["SOFT", "WONK", "opsz"],
});

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const appUrl =
  process.env.NEXT_PUBLIC_APP_URL?.replace(/\/$/, "") ?? "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(appUrl),
  title: {
    default: BRAND.name,
    template: `%s · ${BRAND.name}`,
  },
  description: BRAND.tagline,
  applicationName: BRAND.name,
  referrer: "strict-origin-when-cross-origin",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  viewportFit: "cover",
  themeColor: "#FBF6F2",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${fraunces.variable} ${inter.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="flex min-h-full flex-col bg-bg font-sans text-ink-2">
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
                title: "text-ink font-medium",
                description: "text-muted",
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
