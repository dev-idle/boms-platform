import type { Metadata, Viewport } from "next";
import { Fraunces, Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import { Suspense } from "react";
import { Toaster } from "sonner";

import "./globals.css";
import { ColorModeListener } from "@/components/theme/color-mode-listener";
import { AuthBootstrap } from "@/features/auth/server";
import { COLOR_MODE_INIT_SCRIPT } from "@/lib/theme/color-mode";
import { QueryProvider } from "@/providers";

const fraunces = Fraunces({
  variable: "--font-fraunces",
  subsets: ["latin"],
});

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const appUrl =
  process.env.NEXT_PUBLIC_APP_URL?.replace(/\/$/, "") ?? "http://localhost:3000";

export const metadata: Metadata = {
  metadataBase: new URL(appUrl),
  title: {
    default: "BOMS",
    template: "%s · BOMS",
  },
  description: "Bakery Ordering and Management System",
  applicationName: "BOMS",
  referrer: "strict-origin-when-cross-origin",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  viewportFit: "cover",
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#faf7f5" },
    { media: "(prefers-color-scheme: dark)", color: "#1a141a" },
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${fraunces.variable} ${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      data-mode="light"
      suppressHydrationWarning
    >
      <body className="flex min-h-full flex-col bg-bg font-sans text-foreground">
        <Script id="color-mode-init" strategy="beforeInteractive">
          {COLOR_MODE_INIT_SCRIPT}
        </Script>
        <ColorModeListener />
        <QueryProvider>
          <Suspense
            fallback={
              <div className="flex flex-1 items-center justify-center text-sm text-muted">
                Loading…
              </div>
            }
          >
            <AuthBootstrap>{children}</AuthBootstrap>
          </Suspense>
          <Toaster closeButton richColors position="top-center" />
        </QueryProvider>
      </body>
    </html>
  );
}
