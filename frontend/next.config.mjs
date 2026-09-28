const backendUrl = process.env.BOMS_BACKEND_URL ?? "http://127.0.0.1:8080";

let backendOrigin = "";
try {
  backendOrigin = new URL(backendUrl).origin;
} catch {
  backendOrigin = "";
}

const isProd = process.env.NODE_ENV === "production";

const connectSrc = ["'self'", "https:", "wss:"];
if (backendOrigin && !connectSrc.includes(backendOrigin)) {
  connectSrc.push(backendOrigin);
}

// The realtime socket is dialled straight from the browser, not through the BFF.
// `wss:` already covers it over TLS; plain `ws:` in development needs its origin.
// Production has no local default: its socket is wss://, allowed by scheme.
const realtimeUrl =
  process.env.BOMS_REALTIME_URL ?? (isProd ? null : "ws://localhost:8081/ws");
if (realtimeUrl !== null) {
  const realtimeOrigin = new URL(realtimeUrl).origin;
  if (!connectSrc.includes(realtimeOrigin)) {
    connectSrc.push(realtimeOrigin);
  }
}

const publicAppUrl = process.env.NEXT_PUBLIC_APP_URL;
if (publicAppUrl) {
  try {
    const appOrigin = new URL(publicAppUrl).origin;
    if (!connectSrc.includes(appOrigin)) {
      connectSrc.push(appOrigin);
    }
  } catch {
    /* invalid NEXT_PUBLIC_APP_URL at build time — skip */
  }
}

/**
 * Production drops `'unsafe-eval'` from `script-src` (dev keeps it for Turbopack / tooling).
 * `'unsafe-inline'` stays deliberately: Next.js nonces require dynamic rendering, which would
 * disable Partial Prerendering (`cacheComponents`). XSS defence relies on React escaping, no
 * `dangerouslySetInnerHTML` with user data, and the restrictive directives below.
 */
function buildContentSecurityPolicy() {
  const scriptSrc = isProd
    ? "'self' 'unsafe-inline' blob:"
    : "'self' 'unsafe-inline' 'unsafe-eval' blob:";

  const directives = [
    "default-src 'self'",
    "base-uri 'self'",
    "object-src 'none'",
    "frame-ancestors 'none'",
    "form-action 'self'",
    `script-src ${scriptSrc}`,
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: blob: https:",
    "font-src 'self' data:",
    "worker-src 'self' blob:",
    "manifest-src 'self'",
    `connect-src ${connectSrc.join(" ")}`,
  ];

  if (backendOrigin.startsWith("https://")) {
    directives.push("upgrade-insecure-requests");
  }

  return directives.join("; ");
}

const contentSecurityPolicy = buildContentSecurityPolicy();

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  reactCompiler: true,
  cacheComponents: true,
  poweredByHeader: false,
  compress: true,
  productionBrowserSourceMaps: false,
  experimental: {
    // Instant-navigation validation only where it is achievable: storefront routes opt in with
    // `export const instant = true`. Role areas and sign-in/register wait for the client-side
    // session (the access token is memory-only), so they never prerender and are not opted in.
    instantInsights: {
      validationLevel: "manual-warning",
    },
  },
  // Browser /api/v1/* is proxied by app/api/v1/[...path]/route.ts (BFF) so
  // X-Internal-Secret and Set-Cookie forwarding are reliable. RSC uses lib/api-client.
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-DNS-Prefetch-Control", value: "off" },
          ...(isProd
            ? [
                {
                  key: "Strict-Transport-Security",
                  value: "max-age=63072000; includeSubDomains; preload",
                },
              ]
            : []),
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value:
              "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=(), interest-cohort=()",
          },
          { key: "Cross-Origin-Opener-Policy", value: "same-origin" },
          { key: "Content-Security-Policy", value: contentSecurityPolicy },
        ],
      },
    ];
  },
};

export default nextConfig;
