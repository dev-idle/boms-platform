# BOMS — Architecture Spec

Authoritative reference for backend (Go/Fiber, Hexagonal) and frontend (Next.js App Router, Feature-Sliced). Derived from the current codebase.

---

## 1. High-level architecture

```
┌─────────────────────────┐                ┌────────────────────────────┐
│      Frontend           │  HTTPS         │         Backend            │
│  Next.js 16 + RSC       │  ───────────▶  │  Go 1.26 + Fiber           │
│  Feature-Sliced         │   /api/v1/*    │  Hexagonal (Ports/Adapters)│
└────────────┬────────────┘                └──────────────┬─────────────┘
             │                                            │
       Proxy (Node runtime)                       Postgres (Atlas + sqlc)
       Cookie session check                       Redis (sessions, rate limit, event bus)
       Cross-feature gates                        Asynq (email queue) → SMTP
                                                  PayPal Orders v2 (HTTPS, net/http)
             │                                            │
       Browser ── ws(s)://…/ws?ticket= ──▶  Realtime listener (push-only, own port)
```

---

## 2. Backend — Hexagonal layout

```
backend/
├── cmd/
│   ├── api/main.go              # HTTP entrypoint + composition root
│   ├── worker/main.go           # Worker: sweeps undelivered outbox events, prunes old ones, sends queued email, expires unpaid orders
│   └── genkey/main.go           # Ed25519 key generator CLI
├── db/schema.hcl                # Atlas declarative schema (source of truth)
├── migrations/                  # Atlas versioned SQL (timestamp_*.sql)
├── internal/
│   ├── domain/                  # Entities, value objects, domain errors
│   │   ├── user/                # user, role, audit, errors
│   │   ├── event/               # generic change notice (topic, audience); topics live with their aggregate
│   │   ├── catalog/             # slug, manager audit actions
│   │   ├── category/            # category entity
│   │   ├── product/             # product entity
│   │   ├── profile/             # customer, staff, admin
│   │   ├── session/
│   │   └── store/               # pickup settings, closed days, fixed bakery time zone
│   ├── port/                    # Interfaces (driven + driving)
│   │   ├── user.go, *_profile.go, audit_log.go
│   │   ├── outbox.go            # EventOutbox + EventPublisher
│   │   ├── store.go             # StoreSettingsRepository
│   │   ├── session.go, token.go, password.go, tx.go, health.go
│   ├── usecase/                 # Application services (orchestration)
│   │   ├── auth.go, me.go, admin_user.go, manager_category.go, manager_product.go, catalog.go, readiness.go
│   ├── service/                 # Domain/application services
│   │   ├── profilesvc/          # Role → profile dispatcher
│   │   ├── auditlogger/         # Audit log writer
│   │   ├── eventdispatch/       # Delivers outbox events after commit; sweep + prune for the worker; Fanout
│   │   └── notification/        # Turns committed order events into queued emails
│   ├── handler/v1/              # HTTP handlers (driving adapters)
│   │   ├── auth.go, me.go, admin_user.go, manager_category.go, manager_product.go, catalog.go, health.go
│   ├── adapter/
│   │   ├── repository/
│   │   │   ├── postgres/        # sqlc-backed repos + tx context
│   │   │   │   ├── sql/{query,schema}/
│   │   │   │   └── sqlcgen/     # generated
│   │   │   └── redis/           # sessions, realtime tickets, client
│   │   ├── eventbus/            # Redis Pub/Sub publisher + subscriber; channel names + message shape
│   │   ├── realtime/            # Push-only WebSocket listener: ticket redemption, hub, sockets
│   │   ├── queue/               # Asynq: the email queue (enqueue, one task per event) and its worker handler
│   │   ├── paypal/              # PayPal Orders v2 over net/http: token, create, capture, look up, verify webhooks
│   │   └── email/               # Email templates (HTML + text) and the SMTP mailer
│   ├── infrastructure/          # Pure tech: jwt (EdDSA), crypto (argon2id), logger (zap)
│   ├── middleware/              # auth, ratelimit, cors, security_headers, request_meta
│   ├── shared/                  # ctxmeta, errors, response, utils, validator
│   ├── dto/                     # API request/response shapes (catalog.go, admin_user.go, …)
│   └── bootstrap/               # Composition-root seeding (dev admin)
├── scripts/                     # docker-compose.dev.yml
├── atlas.hcl, sqlc.yaml, Makefile
```

### Dependency rule (strict)

```
domain    ← port    ← usecase / service    ← handler / adapter / cmd
                                            ↑
                              infrastructure (no domain deps)
```

- `domain` depends on nothing.
- `port` depends only on `domain`.
- `usecase` orchestrates ports + domain; never imports adapters.
- `handler` and `adapter` depend on `port` (interfaces) — wired in `cmd/api`.
- `dto` is a leaf: handlers and usecases map domain ↔ dto.

### Naming conventions

| Layer | File pattern | Example |
|-------|-------------|---------|
| Domain entity | `<entity>.go` | `domain/user/user.go` |
| Repo iface | `<entity>.go` | `port/user.go` |
| Postgres repo | `<entity>_repository.go` | `adapter/repository/postgres/user_repository.go` |
| Usecase | `<entity>.go` (singular) | `usecase/admin_user.go` |
| Handler | `<entity>.go` (singular) | `handler/v1/admin_user.go` |
| DTO | `<entity>.go` | `dto/admin_user.go` |
| SQL query | `<entity>.sql` | `sql/query/user.sql` |
| Migration | `YYYYMMDDhhmmss_<name>.sql` | `20260601120000_initial_schema.sql` |

### List pagination (offset, CodeQL-safe)

Canonical implementation: `GET /admin/users` (`handler/v1/admin_user.go`, `usecase/admin_user.go`, `usecase/admin_user_list_page.go`).

| Piece | Location |
|-------|----------|
| Parse query `page` / `page_size` | `shared/utils.ParseQueryInt32` (`strconv.ParseInt`, bit size 32) |
| Generic clamp + offset | `shared/utils.NormalizePageParams`, `PageOffset`, `Int32FromInt64` |
| Feature defaults & max `page_size` | `usecase/*_list_page.go` (int32 const + derived query default string; e.g. admin: 20 / 100) |
| SQL `LIMIT`/`OFFSET` | `port.*Params` as `int32`; usecase returns effective page values for `OKPaginated` meta |
| Page + count together | `usecase/list_with_total.go` — `listWithTotal` runs the page query and its count concurrently (`errgroup`) |

**Rules:** normalize once in usecase (zero-trust); handler does not duplicate clamp; never `strconv.Atoi` → `int` → `int32` for SQL limits. Admin list excludes soft-deleted users (`deleted_at IS NULL`).

**One round trip per page.** Postgres is managed and remote, so a list endpoint
spends its time on round trips, not on queries. Every list goes through
`listWithTotal`, which is read-only by construction: it refuses to run inside a
transaction (`ctxmeta.InTransaction`), because one transaction connection cannot
serve two queries at once. Rows a page needs from another table are joined into
the page query rather than fetched afterwards — the product gallery is a
`LEFT JOIN LATERAL` over the already-paginated rows, not a second query. A list
therefore costs two concurrent connections for the width of one round trip;
`POSTGRES_STATEMENT_TIMEOUT` bounds what a single statement may hold.

**Handler errors:** map usecase failures with `handler/v1/writeMapUsecaseError`; repositories map pgx errors with `mapRepoError`; bind/validation failures use `writeValidationError`.

---

## 3. Frontend — Feature-Sliced layout

```
frontend/src/
├── proxy.ts                     # Next.js proxy (page cookie gate, header strip)
├── app/api/v1/[...path]/route.ts # BFF: browser /api/v1 → Fiber + X-Internal-Secret + Set-Cookie
├── app/                         # Route layer ONLY (thin pages)
│   ├── (public)/                # /, /login, /register, /terms, /privacy, /refund-policy (+ PublicSessionGate layout)
│   ├── (customer)/              # /products, /cart, /orders, /customer/account/*
│   ├── (staff)/                 # /staff/orders, /staff/prep, /staff/account/*
│   ├── (baker)/                 # /baker/production, /baker/account/*
│   ├── (manager)/               # /manager, /manager/categories, /manager/products, /combos, /discount-codes, /account/*
│   └── (admin)/admin/           # /admin, /admin/users, /admin/settings, /admin/account/*
├── features/                    # Feature slices (auth | user | admin | manager | staff | baker | customer | catalog | legal)
│   ├── auth/                    # api/, schemas/, hooks/, components/, lib/, provider/
│   ├── user/                    # api/, schemas/, types/, hooks/, components/
│   ├── admin/                   # api/, schemas/, types/, hooks/, components/
│   ├── manager/                 # catalog CRUD → /api/v1/manager/*
│   ├── staff/                   # order and prep queues → /api/v1/staff/orders/*, /staff/tickets/*
│   ├── baker/                   # kitchen queue → /api/v1/baker/tickets/*
│   ├── customer/                # cart, checkout, orders → /api/v1/cart/*, /orders/*
│   ├── catalog/                 # storefront browse → /api/v1/catalog/*
│   └── legal/                   # policy pages and the consent tick used at sign-up and checkout (no API)
├── components/
│   ├── ui/                      # Primitives (button, input, form, confirm-dialog)
│   └── layouts/                 # dashboard-shell.tsx, staff/manager/baker/admin-shell.tsx
├── lib/
│   ├── api-client.ts, browser-api-client.ts, api-envelope.ts, env.ts, utils.ts
│   ├── validate-next.ts
│   ├── auth/                      # refresh-manager, cross-tab (refresh lock + shared token), session, end-local-session
│   ├── realtime/                  # the tab's push socket: ticket via BFF, reconnect, useLiveQueries
│   ├── server/backend-proxy.ts    # BFF forwarder (browser /api/v1)
│   ├── routing/                   # role-routes.ts, post-auth-destination.ts
│   ├── schemas/auth.ts            # refreshResponseSchema (lib/auth consumer)
│   ├── dal/                       # server-only RSC data access
│   ├── errors/
│   └── validation/
├── stores/                      # Zustand: auth-store only
├── constants/                   # routes.ts, roles.ts, cookies.ts
└── providers/                   # query, theme
```

### Feature-slice anatomy (auth/user/admin all follow this)

```
features/<slice>/
├── api/index.ts                 # Fetch + Zod-validated responses
├── schemas/index.ts             # Zod schemas + inferred types
├── types/index.ts               # Branded ids, derived types
├── hooks/
│   ├── index.ts                 # useXxx mutations/queries
│   └── query-options.ts         # queryKeys + meQueryOptions()
├── components/
│   ├── index.ts                 # Barrel
│   └── <kebab-case>.tsx         # 1 file = 1 component
├── lib/                         # Pure helpers (optional)
├── provider/                    # Context (auth only)
└── index.ts                     # Public surface (barrel)
```

### Shared code placement (no duplicate surfaces)

| Concern | Canonical location |
|---------|------------------|
| Routes / role homes | `constants/routes.ts`, `lib/validate-next.ts` |
| Role route helpers | `lib/routing/role-routes.ts` |
| Post-login / return redirect | `lib/routing/post-auth-destination.ts` |
| Browser API BFF | `app/api/v1/[...path]/route.ts`, `lib/server/backend-proxy.ts` |
| API envelope | `lib/api-envelope.ts` — `parseResponseBody`, `parseApiEnvelope` (browser + server clients) |
| Password complexity (forms) | `lib/validation/password.ts` (`newPasswordZodString`) |
| Internal dashboard chrome | `components/layouts/dashboard-shell.tsx` (+ role shells) |
| Auth refresh / session | `lib/auth/` + `lib/schemas/auth.ts` (`refreshResponseSchema`) |
| Realtime push | `lib/realtime/` (socket, `useLiveQueries`, `LiveIndicator` status) + per slice `<slice>QueryKeysForEvent` in `hooks/query-options.ts` and `<Slice>LiveUpdates` mounted in the role layout |
| RSC auth bootstrap | `features/auth/server.ts` → `provider/auth-bootstrap.tsx` (not `@/features/auth` barrel) |
| Validation messages | `lib/validation/messages.ts` |
| Identity API | `features/user` owns `GET /me`, the personal data export `GET /me/export` and asking for a new confirmation link `POST /me/email-verification` |

**Rules:** no `features/index.ts` meta-barrel; `app/` pages stay thin; RBAC gates are UX — backend enforces roles.

### URL conventions (canonical from `constants/routes.ts`)

| Audience | URLs |
|----------|------|
| Public | `/`, `/login`, `/register`, `/terms`, `/privacy`, `/refund-policy` |
| Customer | `/products`, `/cart`, `/orders`, `/customer/account/{profile,password,delete}` |
| Staff | `/staff/orders`, `/staff/orders/{id}`, `/staff/prep`, `/staff/account/{profile,password}` |
| Baker | `/baker/production`, `/baker/production/:id` (a kitchen ticket), `/baker/account/{profile,password}` |
| Manager | `/manager`, `/manager/categories`, `/manager/products`, `/manager/account/{profile,password}` |
| Admin | `/admin`, `/admin/users`, `/admin/users/{new,[id]}`, `/admin/settings`, `/admin/account/profile` (profile + password) |

**Rule:** one role = one namespace. Each role may only access its own URL prefix (enforced by FE `RoleGate` + post-login redirect). Only `admin` is seeded in development (`bootstrap.EnsureDevAdmin`). No mixing of `/dashboard/*` with `/admin/*`.

### Naming conventions

| Item | Pattern | Example |
|------|---------|---------|
| Page | `page.tsx`, ≤ 10 lines, delegates to feature component | `app/(admin)/admin/users/page.tsx` |
| Layout | `layout.tsx` | `app/(admin)/layout.tsx` |
| Component file | `kebab-case.tsx` | `admin-users-table.tsx` |
| Component export | `PascalCase` | `AdminUsersTable` |
| Hook | `useXxx` | `useMe`, `useUsers` |
| Store | `xxx-store.ts` → `useXxxStore` | `auth-store.ts` |
| Schema | `xxxSchema` + `XxxInput` | `meSchema`, `UpdateSelfProfileInput` |
| Query keys | `<slice>QueryKeys` | `userQueryKeys.me` |

### Import discipline

- Cross-feature imports → through `@/features/<slice>` barrel only (avoid new deep imports; see overview rule).
- **`lib/` must not import `features/*`** — shared contracts live under `lib/schemas/`, `lib/routing/`, etc.
- Role route helpers: `@/lib/routing/role-routes` (canonical).
- Pages import from `@/features/<slice>` or `@/components/ui/*` only.
- No deep imports like `@/features/admin/components/x` from `app/`.
- ESLint blocks `@/lib/api-client` outside `lib/dal/*` (server-only boundary).

---

## 4. Security model

| Threat | Control |
|--------|---------|
| Token theft | EdDSA-signed short-lived access JWT + HttpOnly refresh cookie (`Path=/`; consumed only on `/api/v1/auth/*`) |
| Unauthenticated page access | Next.js `proxy.ts` coarse gate: protected HTML routes require refresh cookie **presence**; cookie `Path` must be `/` (`middleware.AuthCookiePath`) so the browser sends it on `/admin`, `/products`, etc. |
| Session hijack | Redis session store; bearer + cookie hybrid logout; `RequireAuthWithSession` for mutating routes. Ending every session of an account (a new password — changed, reset or set by an administrator — disabling it, or an administrator's revoke) also raises `users.session_version` in the same write; a refresh token carries the version read with the password hash at sign-in, and refresh drops a session whose version is older. The store sweep (SCAN) can miss a session being refreshed or created at that moment; the version cannot, so such a session never refreshes again — it lasts until its access token expires (`JWT_ACCESS_TTL`), plus `REALTIME_MAX_LIFETIME` for a realtime socket opened in that time |
| Password attack | Argon2id (params from config); timing-safe dummy hash on login |
| Replay | request-id propagation; refresh rotation |
| CSRF | SameSite=Lax cookie; the BFF refuses a write (any method but GET/HEAD/OPTIONS) whose `Sec-Fetch-Site` is present and not `same-origin` with 403 `forbidden` (requests without it, such as server calls, pass), so another site cannot post a forged sign-in or borrow visitors' browsers against per-IP limits; internal proxy secret on inbound headers, sanitize `X-User-Role`, `X-Request-ID`, `X-Auth-Hint` |
| Bruteforce | Redis-backed rate limit per IP (login/refresh/logout) + per-user admin writes (30/min), manager catalog writes (30/min, `RATE_LIMIT_REDIS_MANAGER_WRITE_*`), order mutations — checkout, starting and capturing a payment, staff order status, ticket status and ticket moves (20/min, `RATE_LIMIT_REDIS_ORDER_WRITE_*`) + self-service account writes — `PATCH /me`, `PATCH /me/password`, `DELETE /me` (10/min, `RATE_LIMIT_REDIS_SELF_WRITE_*`) + discount code attempts per customer — `PUT /cart/discount` (10/15 min, `RATE_LIMIT_REDIS_DISCOUNT_ATTEMPT_*`) + manager Cloudinary signatures (20/min, `RATE_LIMIT_REDIS_MANAGER_MEDIA_*`) + realtime tickets per user (60/min, `RATE_LIMIT_REDIS_REALTIME_TICKET_*`) + personal data exports per user — `GET /me/export` (5/hour, `RATE_LIMIT_REDIS_DATA_EXPORT_*`) + emailed links per IP — `POST /auth/verify-email`, `POST /auth/password-reset/confirm` (10/15 min, fails closed, `RATE_LIMIT_REDIS_AUTH_LINK_*`) + reset requests per IP — `POST /auth/password-reset/request` (5/15 min, fails closed, `RATE_LIMIT_REDIS_PASSWORD_RESET_*`; each account is also sent at most 3 links an hour, whoever asks, `RATE_LIMIT_REDIS_PASSWORD_RESET_ACCOUNT_*`) + new confirmation links per user — `POST /me/email-verification` (3/hour, `RATE_LIMIT_REDIS_VERIFICATION_RESEND_*`) + PayPal webhook deliveries per IP — `POST /payments/paypal/webhook` (60/min, `RATE_LIMIT_REDIS_PAYMENT_WEBHOOK_*`) |
| RBAC | `RequireRole(Admin)` on `/admin/*`; admin can't modify self; staff self-update only fills `full_name`, `phone` |
| Forced password change | `must_change_password` flag → `RequirePasswordChanged` middleware blocks all routes except `/me` GET and `/me/password` PATCH |
| Audit | All admin mutations write to `audit_logs` with actor/target/before/after |
| Soft delete | `users.deleted_at` (no hard delete from app) |
| Realtime socket | The only public surface besides `/health` and `/ready`, on its own port and path (`/ws`). Admission refuses a missing or unlisted `Origin`, a malformed token and anything over the per-address or per-process redeem rate before touching Redis, then needs a single-use ticket issued through the BFF to a signed-in session (32 random bytes, stored hashed, 30 s, redeemed with `GETDEL`) whose session still exists, and a free slot under the per-user and per-process caps. Handshakes are bounded by `REALTIME_HANDSHAKE_TIMEOUT`. Channels are chosen from the ticket, never from the browser; the socket carries change notices only and refuses any data frame from the browser |
| Multi-tab sessions | Tabs share one refresh cookie and the API revokes every session when a refresh token is spent twice, so tabs refresh one at a time (Web Locks) and pass the new access token to the others (`BroadcastChannel`, same origin, adopted only by tabs of the same account) |
| Inbound | Strip `x-internal-secret`, `x-user-role`, `x-auth-hint` and the forwarding headers (`x-forwarded-for`, `x-real-ip`, `forwarded`, `x-client-ip`) from client requests in proxy |
| Landing flash | The access token is memory-only, so a returning visitor is anonymous until the session is restored. Fiber issues a role cookie beside the session cookie (`COOKIE_ROLE_NAME`) and `proxy.ts` lands them at the edge (`lib/routing/signed-in-landing.ts`). A **navigation hint only** — never authorization: every byte of data stays behind the API session check, and a forged or stale role reaches a page whose gate sends it back. **Document navigations only** (`Sec-Fetch-Dest: document`, never `/api/*`): on a client navigation the hint would fight the gate that knows the real session, and the two would bounce a visitor between them — including away from the sign-in page, exactly when the session cannot be restored. The redirect carries `Cache-Control: no-store, private` and `Vary: Cookie`, because a cookie decided it |
| Visitor identity | The API only ever sees the BFF. `proxy.ts` resolves the visitor from what the hosting edge reported (`resolveClientIp`) before it drops the forwarding headers, stamps the checked address as `X-Client-IP`, and the BFF forwards that stamp (`stampedClientIp`, `lib/server/client-ip.ts`); `middleware.ClientIP` reads it, falling back to the socket address when it is absent or malformed. Rate-limit buckets and `audit_logs.ip` would otherwise all be the BFF |
| Customer data by role | Each role gets only the customer data its work needs, decided in the usecase DTO mappers: staff get name and email on the order list and the phone as well on one order (`StaffOrderCustomerResponse`) to coordinate its pickup — a list page never hands out every customer's number; the kitchen and the prep queue get the display name only (`TicketCustomerResponse`) — a station never contacts a customer |
| Abandoned request | `HTTP_REQUEST_TIMEOUT` deadlines the handler context (below `HTTP_WRITE_TIMEOUT`), `POSTGRES_STATEMENT_TIMEOUT` caps one statement — a caller that walks away cannot hold database connections |
| Payments | The buyer approves on PayPal's own page — a redirect, no PayPal script or frame, so the CSP stays closed and BOMS never sees card or account details; the approve URL the browser follows must be an https `paypal.com` address (Zod). The amount is the order total the server priced, and a capture is recorded only when its amount and currency equal the order's — otherwise nothing changes and the error is logged. `PAYPAL_CLIENT_SECRET` stays on the API and the worker (`PAYPAL_MODE=live` required in production). The webhook is the one `/api/v1` route without a session: PayPal posts through the BFF, and nothing is read from a delivery until PayPal confirms its signature for `PAYPAL_WEBHOOK_ID` (required in staging/production; without it every delivery is refused); a delivery can only move a payment forward along the path a capture takes, so a replay changes nothing |

### Product images (Cloudinary)

Signed upload keeps `CLOUDINARY_API_SECRET` on the API only.

| Step | Owner |
|------|-------|
| Manager requests signature | `GET /api/v1/manager/media/cloudinary-signature` (`manager` role, session, per-user rate limit) |
| Browser uploads file | Direct `POST` to `https://api.cloudinary.com/v1_1/{cloud}/image/upload` |
| Persist URLs | `product_images` table (max 5 per product, `sort_order` 0–4) stores Cloudinary `secure_url` values |
| Storefront image URLs | `catalogProductImageUrl()` / gallery on detail; cards use first image |

**Validation (defense in depth):**

| Layer | Rule |
|-------|------|
| Signed params | `folder`, `allowed_formats`, `unique_filename`, `timestamp` |
| Size cap | API returns `max_bytes` (5 MiB) for client-side checks; FE validates `file.size` before upload and response `bytes` when present |
| Persist | BE `IsCloudinaryDeliveryURLInFolder` on create/update when Cloudinary is enabled |
| FE submit | Zod `productImageUrlsSchema` (max 5) mirrors folder + cloud rules |

When Cloudinary env is set on both sides, product create/update rejects URLs outside the configured upload folder. When unset (development only), managers may use any HTTPS image URL.

**Production:** `cloudinary.cloud_name`, `api_key`, and `api_secret` are required when `app.env` is `staging` or `production`.

Env (must match):

| Backend | Frontend |
|---------|----------|
| `CLOUDINARY_CLOUD_NAME` | `NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME` |
| `CLOUDINARY_UPLOAD_FOLDER` (default `boms/products`) | `NEXT_PUBLIC_CLOUDINARY_UPLOAD_FOLDER` when overriding |

URL path parsing for folder checks is duplicated in `backend/internal/domain/media` and `frontend/src/lib/cloudinary/config` — keep tests in sync when changing either.

---

## 5. Performance principles

- **Backend:** prepared statements via sqlc, paginated queries (max 100), single JOIN for admin user listing, `GET /me` JWT-only (no Redis), session meta cached in Fiber Locals.
- **JWT:** stateless access JWT for read routes (`GET /me`); Redis session only on writes — limits Redis QPS.
- **Frontend:** Turbopack build, RSC + Partial Prerendering, route-group code-split per role, Zod parse only at boundary.
- **Polling avoidance:** TanStack Query cache + invalidation on mutation success and on pushed change notices (§6 Realtime push) — no polling.
- **Concurrency:** Postgres transactions via `TxManager` for multi-table writes (auth register; admin role and profile edits, which lock the account row and edit the shared staff profile in place; phone writes, which take a per-number advisory lock before the duplicate check; enable, which restores the account and releases its phone if an active account took it meanwhile; checkout and order status moves, which write their outbox event in the same transaction).

---

## 6. Module ownership (canonical sources of truth)

| Concern | Owner |
|---------|-------|
| Session identity (`/me`) | `features/user` (FE) + `usecase/me` (BE) |
| Customer policies and personal data | `features/legal` + `features/user` (FE) + `domain/policy` + `usecase/data_export` (BE) — `/terms`, `/privacy`, `/refund-policy`, `GET /me/export` |
| Auth (login/register/logout) | `features/auth` (FE) + `usecase/auth` (BE) |
| Admin user CRUD | `features/admin` (FE) + `usecase/admin_user` (BE) |
| Store settings (pickup rules) | `features/admin` settings + `features/customer` checkout panel (FE) + `domain/store` + `usecase/admin_store_settings` + `usecase/store` (BE) — `/admin/settings`, `/admin/closed-dates/*`, public `/store/pickup-rules` and `/store/pickup-slots` |
| Manager catalog CRUD | `features/manager` (FE) + `usecase/manager_category` + `usecase/manager_product` + `manager_combo` + `manager_discount_code` (BE) |
| Product images (Cloudinary) | `lib/cloudinary/*` + `components/ui/catalog-image-list-field` (FE) + `usecase/manager_media` + `service/cloudinary` (BE) |
| Storefront catalog browse | `features/catalog` (FE) + `usecase/catalog` (BE) — API path `/catalog/*` |
| Customer cart & checkout | `features/customer` (FE) + `usecase/cart` + `usecase/order` (BE) — `/cart/*`, `/orders/*` (session + server pricing; **pickup-only**, see below) |
| Payments (PayPal) | `features/customer` order page (FE) + `domain/payment` + `usecase/payment` + `adapter/paypal` + `adapter/repository/postgres/payment_repository.go` + the expiry job in `cmd/worker` (BE) — `/orders/:id/payment`, `/orders/:id/payment/capture`, `/payments/paypal/webhook` |
| Staff order queue | `features/staff` (FE) + `usecase/staff_order` (BE) — `/staff/orders/*` (list, detail, status transitions) |
| Staff prep queue | `features/staff` (FE) + `usecase/staff_ticket` (BE) — `/staff/tickets/*` (counter tickets, ticket status, move a ticket to the other station) |
| Baker kitchen queue | `features/baker` (FE) + `usecase/baker_ticket` (BE) — `/baker/tickets/*` (kitchen tickets, ticket detail, ticket status) |
| Audit logs | `service/auditlogger` (BE only) |
| Event delivery | `domain/event` + `port/outbox.go` + `adapter/repository/postgres/outbox_repository.go` + `adapter/eventbus` + `service/eventdispatch` + `bootstrap.EventPublisher` + `cmd/worker` (BE only) |
| Order emails | `domain/order/notice.go` + `service/notification` + `adapter/queue` + `usecase/order_email` + `adapter/email` + `cmd/worker` (BE only) |
| Account links (confirm email, reset password) | `domain/account` + `usecase/{email_verification,password_reset,account_email}` + `adapter/repository/postgres/user_token_repository.go` + `adapter/email` (FE: `features/auth` link pages, `features/user` `EmailVerificationNotice`) |
| Realtime push | `lib/realtime` + slice live updates (FE) + `usecase/realtime` + `adapter/realtime` + `adapter/eventbus` subscriber (BE) |
| Profile entity dispatch | `service/profilesvc` (BE) |
| Routes table | `constants/routes.ts` (FE) |
| Roles enum | `constants/roles.ts` (FE) + `domain/user/role.go` (BE) |
| Validation messages | `lib/validation/` (FE) + `shared/validator/` (BE) |

### Event delivery (transactional outbox)

A change other people must see writes an `outbox_events` row **in the same transaction** as the change, so an event exists exactly when the change committed.

| Step | Where |
|------|-------|
| Record | usecase calls `EventOutbox.Add(txCtx, event)` inside `TxManager.WithTx`; outside a transaction it is refused |
| Deliver | after the commit, `Pool.WithTx` hands the transaction's events to `eventdispatch.Dispatcher.AfterCommit`, which publishes them in the background (bounded by `OUTBOX_DISPATCH_TIMEOUT`) and then, on a fresh deadline of the same length, marks them published or records the failure on the row — a publish that timed out is still written down. A rollback publishes nothing. The API waits for these deliveries on shutdown |
| Recover | `cmd/worker` sweeps rows still unpublished `OUTBOX_SWEEP_GRACE` (longer than twice the dispatch timeout) after they were written — at start and then every `OUTBOX_SWEEP_INTERVAL`, draining full batches at once. `FOR UPDATE SKIP LOCKED` means two workers never hold the same row at once; a sweep that published but could not mark rolls back and sends again. The worker validates only the database, Redis, outbox, order, PayPal and mail settings (`config.LoadWorker`) |
| Retain | published rows are deleted after `OUTBOX_RETENTION` — delivery records, not business data, so no soft delete |
| Clock | `created_at` defaults to `clock_timestamp()` and every age (sweep grace, retention) is measured in SQL, so the API and the worker never compare two hosts' clocks |

Delivery is **at least once** and **unordered across transactions** (each commit is delivered on its own): an event may arrive twice or after a later one, and subscribers use it only as a hint to refetch through the API. The bus is Redis Pub/Sub (`adapter/eventbus`): channel `boms:events:user:{id}` per user, `boms:events:role:{role}` per role and `boms:events:public` for notices with nothing private in them, message `{id, type, at, data}` — identifiers and labels, never the changed record.

| Topic | Written by | Audience | Data |
|-------|------------|----------|------|
| `order.created` | checkout | the customer — the bakery hears of an order once it is paid | `order_id`, `status` |
| `order.status_changed` | staff status moves, the status a ticket move derives, a captured payment (`awaiting_payment` → `confirmed`) and expiry | the customer · staff unless the order was never paid (`Status.VisibleToStaff`) · baker only when the order enters, leaves or moves within the statuses bakers see (`Status.VisibleToBaker`) | `order_id`, `from`, `status` |
| `ticket.changed` | a station starting or finishing a ticket, a staff move, and cancelling or expiring its order | the customer · staff unless the order was never paid · baker only when bakers see the order (`Status.VisibleToBaker`; for a cancellation, the status it left) — a kitchen ticket shows where the order's other tickets stand | `ticket_id`, `order_id`, `station`, `status` |
| `settings.updated` | admin settings and closed-day changes | everyone with a page open (public channel) | none |
| `slots.changed` | a checkout that takes a slot's last place, and any cancellation or expiry | everyone with a page open (public channel) | `date` (the bakery day of the slot) |

Topics and their audiences live with the aggregate that raises them (`domain/order/event.go`, `domain/store/event.go`), as audit actions do.

The API, and the worker for the orders it expires, deliver after commit; run `cmd/worker` beside the API (`make run-worker`) to recover what a delivery missed and to send email.

Every delivery — the API's after commit and the worker's sweep — goes to the same destinations, `bootstrap.EventPublisher`: the realtime bus, then the email queue (`eventdispatch.Fanout`). A delivery that fails at either leaves the event pending and hands it to both again, so each takes an event twice.

### Order emails

A customer gets an email at the moments they act on: the order was paid and received (`order.status_changed` from `awaiting_payment` to `confirmed`), it is ready to collect (`order.status_changed` to `ready`), it was cancelled, or it expired unpaid. Steps inside the bakery (confirmed, in production) show on the order page instead. `domain/order.NoticeFor` decides from the event.

| Step | Where |
|------|-------|
| Queue | `service/notification.OrderEmails` is a publisher on the outbox delivery path: for each event that calls for a notice it enqueues an Asynq task (`adapter/queue`, queue `email`) whose **task id is the event id**. A second enqueue for the same event is refused by Asynq and counts as done; sent tasks are kept 24 h so a late redelivery still finds them. The row is marked published only once the task is in the queue, which Redis keeps across restarts (append-only file) |
| Send | `cmd/worker` runs the Asynq server (`MAIL_CONCURRENCY` at once). `usecase/order_email` reads the order as it is now (`StaffGetByID`, which joins active accounts only) and skips the email when the customer closed their account, has not confirmed their address, or the order has moved past the notice (`Notice.StillApplies` — no "ready" for an order already collected). `adapter/email` writes it — HTML in the bakery's style with a plain-text part, pickup time in bakery time, a receipt on the "received" email, a link to the order — and sends it over SMTP (`MAIL_*`; STARTTLS or implicit TLS, credentials and an https `APP_SITE_URL` required in staging/production), from `MAIL_FROM_ADDRESS` with replies to `MAIL_REPLY_TO` |
| Retry | a failed or dropped connection, a timeout or a "try again later" (4xx) is retried with Asynq's backoff, up to 8 times (about an hour and a half); only a 5xx refusal (unknown address) or an email that cannot be written is final — archived in Asynq for inspection and logged as an error. A message the server accepted is never sent again, even if hanging up fails |
| Redis | the queue lives in Redis, so a deployed Redis must keep its data: append-only persistence on, and `maxmemory-policy noeviction` — an evicted task is an email lost. The dev compose file runs Redis with the append-only file |
| Once | the queue keeps one task per event, and Asynq runs a task at least once: a worker stopped after the server accepted a message but before recording it sends that email again. That is the one rare duplicate; nothing else sends twice |
| Privacy | the task holds the event id, order id and notice — never the address, which is read at send time. Send errors name the SMTP stage and reply code, not the recipient, because they are logged and kept on the task (`adapter/email/smtp.go`) |
| Dev | Mailpit (`scripts/docker-compose.dev.yml`) catches every message at http://localhost:8025; the worker's defaults point at it |

The bakery's name and contact details in the emails match the storefront's; both are tested against `contracts/brand.json`.

### Realtime push (WebSocket)

A Route Handler cannot accept a WebSocket upgrade, so the socket cannot pass through the BFF. The BFF stays the only way to *obtain* access; the socket gets its own small, push-only listener in `cmd/api` (second `http.Server` on `REALTIME_ADDR`, `github.com/coder/websocket`, `/ws` only).

```
browser ── POST /api/v1/realtime/tickets (BFF, signed-in session) ──▶ { ticket, url }
browser ── GET url?ticket=… (WebSocket) ──▶ realtime listener ── redeem ticket, check Origin, per-user cap
Redis boms:events:* ── one pattern subscription per API process ──▶ hub ──▶ sockets on the ticket's channels
socket message ──▶ TanStack Query invalidation ──▶ refetch through the BFF (authorization stays on REST)
```

| Concern | Rule |
|---------|------|
| Channels | every ticket hears its own `boms:events:user:{id}` and `boms:events:public`; staff and baker also hear their role channel. Nothing private travels on a channel a customer shares |
| Lifetime | the socket re-checks its session every `REALTIME_SESSION_CHECK_INTERVAL` and closes with `4001` once it is gone (logout, revocation, token refresh — every refresh rotates the session) or after `REALTIME_MAX_LIFETIME`; `4001` tells the tab to fetch a new ticket at once. A Redis error is tolerated for two checks in a row; the third closes the socket (`1013`) |
| Liveness | ping every `REALTIME_PING_INTERVAL`; each socket has `REALTIME_SEND_BUFFER` queued events and is closed (`1013`) when it falls behind, instead of slowing the hub |
| Browser | one socket per tab, opened by the first `<Slice>LiveUpdates` and closed a second after the last unmounts; reconnect waits grow to 30 s with jitter and are skipped when the tab returns or the network comes back; every (re)connect refetches what the page shows, since events sent meanwhile are gone |
| Shutdown | the listener stops admitting, open sockets get `1001`, then the bus subscription ends — before the API waits for its outbox deliveries |
| Deploy | browsers reach the listener directly: expose it as `wss://<site>/ws` (edge proxy, TLS), set `REALTIME_PUBLIC_URL` to that URL, `REALTIME_ALLOWED_ORIGINS` to the site origin exactly as browsers send it, `REALTIME_TRUSTED_PROXIES` to the edge so admission meters the real client address, and `BOMS_REALTIME_URL` on the frontend so CSP `connect-src` allows it in development (`wss:` is allowed by scheme). The `?ticket=` query must not be logged by the proxy |

### Fulfillment model (pickup-only — no Address module)

BOMS is a **bakery pickup** flow, not delivery or shipping.

| Topic | Spec (canonical) |
|-------|------------------|
| **Fulfillment** | Customer orders for **in-store / counter pickup** at the bakery. |
| **No Address module** | No `addresses` table, no shipping/delivery address on profile or orders, no geocoding, no carrier integration. **Do not add** unless this document is updated first. |
| **Customer profile** | `customer_profiles`: `display_name`, `phone` (+ account `email` on `users`). Phone is contact info for pickup coordination — **not** a delivery address. |
| **Checkout / orders** | `orders`: `code`, `order_type`, pricing, discount snapshot, `status`, required `pickup_at` at checkout, held to the Admin pickup rules below, line items with `configuration` jsonb — **no** shipping/delivery address fields. |
| **Storefront `BRAND.addressLine`** | Static marketing copy for footer “Visit us” (`constants/brand.ts`) — the **bakery location**, not per-customer data. |
| **Marketing copy** | UI may say “pickup” but must not imply saved delivery addresses or ship-to-door unless a feature is implemented. |

**Pickup rules (Admin settings, `store_settings` + `store_closed_dates`):** opening hours, slot length and capacity, pre-order notice, the counter's instant preparation time, booking window and time to pay live in one settings row (migration defaults: 08:00–18:00, 30-minute slots of 10 orders, 2 h, 20 min, 14 days, 15 min), plus closed days with a reason customers see. Checkout reads them per request — one primary-key row and the closed days inside the window, fetched together — so an edit applies to the next checkout at once; open carts refresh through `settings.updated`. Refusals are specific: `pickup_too_soon`, `pickup_too_far`, `pickup_closed_day`, `pickup_outside_hours`, `pickup_off_slot`, `pickup_slot_full`, `pickup_day_limit`.

**Order type and slots:** categories name their station (`kitchen` makes to order, `counter` sells ready-made) and products their own notice (`lead_time_minutes`). Checkout decides the order type from the locked cart's items: counter items only, collected the day they are ordered, make an **instant** order that waits the counter's preparation time; anything else is a **pre-order** that waits the pre-order notice — either waits the longest notice an item needs (combos count by what they hold). Pickups start on slot marks from opening; checkout holds its slot with a transaction-scoped advisory lock and refuses it once the slot holds as many orders as it takes — every order neither cancelled nor expired whose pickup falls in the slot, so orders booked on an older grid still count. One customer may hold at most three orders for a bakery day (`order.MaxOrdersPerCustomerPerDay`), so no single account can fill the day's slots. `GET /store/pickup-slots?date=` lists a day's slots marked full, the cart reports what its items need (`fulfillment`), and the picker offers the slots that suit them; `slots.changed` goes out when a checkout fills a slot or an order is cancelled or expires, so it tells no more than the slot list shows. The time zone is fixed (Asia/Ho_Chi_Minh, no DST): the bakery does not move, and every stored hour is read in it.

**Order status (schema v2):** `awaiting_payment` → `confirmed` → `in_production` → `ready` → `fulfilled` | `cancelled`; an order not paid in time becomes `expired`. `pending` is an order placed before online payment, which staff still confirm.

| Role | Allowed transitions |
|------|---------------------|
| **Customer** (paying) | `awaiting_payment`→`confirmed` when PayPal takes the money |
| **System** (worker) | `awaiting_payment`→`expired` once `payment_due_at` has passed; recorded with no actor |
| **Staff** | `pending`→`confirmed`\|`cancelled`; `confirmed`\|`in_production`→`cancelled`; `ready`→`fulfilled`\|`cancelled` |
| **Tickets** (derived) | `confirmed`→`in_production` when the first ticket starts; `in_production`→`ready` when every ticket is ready |

**Staff workflow:** counter confirm/cancel, its own prep queue, and handoff when `ready`; the kitchen works its own tickets.

**Payment (PayPal Orders v2, USD, intent CAPTURE, over `net/http`):** checkout places an order `awaiting_payment` that holds its slot and its discount use until `payment_due_at` — the transaction clock plus the admin's time to pay (5–120 min); an order with nothing to pay is confirmed at once. Checkout requires an `Idempotency-Key` (UUID) kept per customer (`orders.checkout_key`, unique with `user_id`), so a retried checkout answers with the order the first attempt placed — also when both run at once, since the second waits on the cart row and finds it emptied. `POST /orders/:id/payment` creates the PayPal order — the order total, `custom_id` = order id, return and cancel URLs on `APP_SITE_URL` — or reuses the one still open, and answers its approve URL; PayPal sends the buyer back to `/orders/:id?paypal=approved`, where the page calls `POST /orders/:id/payment/capture` once. Paying is refused once `payment_due_at` has passed, and an order expires only two minutes after it, so a capture started in time finishes first; it captures with `PayPal-Request-Id` = `capture-<paypal order id>` so a retry gets the first answer, then records the capture and confirms the order in one transaction. PayPal is asked for immediate funding only (no eCheck), so a capture it holds for review (`pending`, the order still awaiting payment) is rare; a declined card lets the buyer approve again with another funding source, and a capture PayPal refuses (`DECLINED` or `FAILED`, recorded `denied`) lets them start again with a new PayPal order. `payments` keeps one row per order (`created` → `pending` → `captured` | `denied`), every move guarded in SQL, so the return page, the webhook and the worker may report the same capture in any order and it is recorded once. The webhook is the backup for a buyer who never comes back; the worker is the last: every `ORDER_EXPIRY_INTERVAL` it takes up to 50 orders more than two minutes overdue and asks PayPal about each — money taken confirms the order, a capture under review leaves it waiting, anything else (including a PayPal order PayPal no longer has) expires it (guarded on status and due time), cancels its tickets, frees its slot, gives back its discount use and emails the customer. A capture for an order that closed meanwhile is kept and logged (`payment_captured_for_closed_order`) so the money can be returned. Unpaid and expired orders are the customer's alone: staff lists and order pages, the kitchen's queue and the staff and kitchen channels see an order once it is paid. A discount code's use is taken at checkout under the code's row lock (`max_uses`), and only then are the customer's own uses counted (`max_uses_per_customer`: their orders neither cancelled nor expired), so two checkouts never take the last use twice.

**Tickets:** production works on `order_tickets`, not on the priced receipt lines. Checkout splits an order into one ticket per station (`order_ticket_items`; a combo arrives as its products, each at its category's station). A ticket goes `queued` → `in_progress` → `ready`, or `cancelled` with its order; the kitchen moves kitchen tickets (`/baker/tickets/*`), the counter its own (`/staff/tickets/*`), and only once the order is accepted (`confirmed` or `in_production`). Every ticket move locks the order row first (`OrderRepository.LockForUpdate`) and derives the order status in the same transaction (`order.DeriveStatus`): the first ticket started puts the order `in_production`, the last one ready makes it `ready`, recorded in `order_status_events` as moved by whoever moved the ticket. With the lock taken first, two stations finishing at once make the order ready once. Staff may move a ticket nobody started to the other station — one ticket per station per order, audited. Cancelling an order cancels its tickets in the same transaction. Order details carry the tickets: customers and the kitchen see where each station stands, staff also see what each makes.

**Policies and personal data:** the terms of sale, privacy policy and refund policy are one versioned set (`domain/policy.TermsVersion`, `constants/policies.ts`, both tested against `contracts/terms-version.json`). Registration and checkout send the version the customer ticked and are refused with `terms_not_accepted` (422) unless it is the current one, so a page left open across a policy change does not count as agreement; the version and the instant (the transaction clock, like `created_at`) are recorded on `users` and on each `orders` row (`terms_version`, `terms_accepted_at`, both set or neither). `GET /me/export` (session, password changed, rate limited, `Cache-Control: no-store`) returns everything held about the signed-in person: account and profile as `GET /me` shapes them, the version accepted at sign-up, their live sign-in sessions (when, network and browser — never a session's id or token), the changes recorded to the account (`audit_logs` about them; the network and browser only for changes they made themselves), and for a customer their cart and every order with its lines and history. Orders and changes are read 100 at a time by keyset on `(created_at, id)`, so a row written while the export reads never repeats or goes missing, with each order page's lines and histories in one round trip. The frontend saves it as a JSON file without dropping any field.

**Confirming an address and resetting a password:** a customer who signs up is unconfirmed (`users.email_verified_at` null); accounts an administrator creates count as confirmed, and every account from before this existed was backfilled as confirmed. Checkout refuses an unconfirmed account with `email_not_verified` (422), and order emails skip one, because order updates go to that address. Both flows use single-use links (`user_tokens`: at most one per user and purpose, SHA-256 of a 256-bit token, expiry on the database clock — 48 h to confirm, 30 min to reset). The link is made by the worker when it sends the email (`usecase/account_email`), from a task holding ids only, so a token never sits in an event, the realtime bus or the queue; it rides in the URL fragment (`/verify-email#token=…`, `/reset-password#token=…`), which a browser never sends to a server, and the page clears it from the address bar. Redeeming deletes the row in the statement that reads it, so a link works once, and a link of a closed account opens nothing; every failure is the same `invalid_link` (422). Registration asks for the confirmation email in its own transaction (`account.verification_requested`, an outbox event with no audience), and `POST /me/email-verification` asks again. `POST /auth/password-reset/request` answers 202 whatever the address, and only an open account gets a link (`account.password_reset_requested`). `POST /auth/password-reset/confirm` sets the password, lifts a required change, confirms the address (the link reached the inbox), records `me.reset_password` and ends every session inside the transaction — if they cannot end, nothing changes and the link still works — then sweeps again after the commit. `POST /auth/verify-email` records `me.verified_email`. Neither link needs a session: they are often opened on another device. `POST /me/email-verification` for an address already confirmed sends nothing and answers `email_already_verified` (409), so the page can say so. A new password of any kind — changed, reset from a link or set by an administrator — and disabling the account delete every link the account was emailed, in the same transaction, so a reset link sent earlier cannot undo the change; each of these takes the link rows before the account row, the order redeeming a link takes them. The emails travel like order emails (below): `service/notification.AccountEmails` enqueues an `email:account` task per event (task id = event id), retried up to 5 times and kept 1 h after sending.

**Erasing an account:** `DELETE /me` is a customer's right to erasure (`usecase/account_erasure`), not a soft delete. It cannot be undone, so it takes the current password again (`{"password"}`; a wrong one is `invalid_credentials`, 401). It is refused with `account_has_open_orders` (422) while any of their orders is neither fulfilled, cancelled nor expired — the bakery still has to make or hand it over. It locks the cart row and then the account row before that check, the order checkout takes them in: checkout holds the account with `FOR SHARE` while it places an order, so an erasure waits for that order and sees it. A checkout that starts during an erasure waits on the cart and finds it emptied; one with a cart filled after the erasure, or during an admin disable, finds the account closed (`me_not_found`). `PATCH /me` holds the account the same way and records its audit entry in its own transaction, so a profile change never lands on, or outlives the scrub of, an erased account. In one transaction it then empties the cart, clears the profile's name and phone, turns the email into `erased-<id>@erased.invalid` (freeing the address, RFC 2606), makes the password unmatchable, closes the account (`deleted_at` and `erased_at`), records `me.erased_account` (the record commits with the erasure or not at all), deletes its emailed links, and scrubs the audit trail: entries about the account and its profile lose their before/after values and the account's own entries lose their network and browser, while what happened, when and by which role stays. Every session ends after the commit; if that fails, asking again finishes it — for an erased account `DELETE /me` only ends the remaining sessions and answers 204. The row itself stays, so the customer's orders remain anonymous sales records, which accounting law requires. An erased account never comes back: admins see it as Erased with Enable greyed, `Enable` refuses it with `account_erased`, and the restore query ignores it. Staff accounts are closed by an administrator, not erased by their holder.

**Order code and history:** every order carries `CH-YYMMDD-NNN` — the bakery day it was placed and its number that day. Checkout takes the number from `order_day_counters` with an upsert inside its transaction, on the bakery day of the transaction clock (the instant `created_at` records); the counter row stays locked until commit, so numbers are unique and gap-free per day. `order_status_events` records each status an order enters, when, and who moved it (actor id and role — neither when the system expired an unpaid order), written in the transaction that makes the move — checkout writes the first entry; payment, staff transitions, ticket moves and expiry the rest. Order details return it as `timeline` (customers: status and time; staff: also the actor role, null for the system); the kitchen gets the code only. Customers filter their history by `status` and by the bakery days orders were placed on (`from`, `to`, both inclusive).

---

### Shared contracts (`contracts/`)

Rules enforced on **both** sides are held to one fixture at the repo root, owned
by neither side:

```
contracts/
├── README.md                    # what belongs here, how to add a case
└── vietnam-phone-cases.json     # Vietnam mobile rule — accepted, stored, rejected, displayed
```

`backend/internal/shared/utils/phone_test.go` and
`frontend/src/lib/validation/phone.test.ts` read the same file, and both CI
workflows watch `contracts/**`, so a change to one implementation that the other
does not follow fails the build. Cases go in the fixture, never in either test.

A fixture belongs here only when Go and TypeScript both enforce the rule and must
agree exactly. A rule that lives on one side stays with that side; codes mirrored
across the boundary (`apperrors` → `ApiErrorCode`) are covered by
`internal/shared/errors/contract_test.go` instead.

---

## 7. Verification gates

```bash
# Backend
cd backend
go vet ./...
go test ./...           # all green
go build ./...          # all packages compile
golangci-lint run --timeout=4m
make sqlc-check         # generated code in sync (CI)

# Frontend
cd frontend
pnpm typecheck          # tsc --noEmit
pnpm lint
pnpm test               # Vitest — validate-next, role-routes, …
pnpm build              # production build
```

CI must run backend tests + frontend typecheck, lint, test, and build. Production deploys block on failure.

### Production guardrails (enforced in code + CI)

| Area | Backend | Frontend |
|------|---------|----------|
| **Security** | `RequireRole` per route; `X-Internal-Secret`; soft-delete filters; session revoke on role/disable | `proxy.ts` strips spoofed headers; `validateNext` / `validateNextForRole`; HttpOnly refresh; access token in memory only |
| **Performance** | Pagination caps in usecase; `ParseQueryInt32`; `GET /me` no Redis; Locals session meta | 25s fetch timeout; one refresh at a time across tabs; `cache: no-store` on API |
| **Clean boundaries** | Handler → usecase → port; `mapRepoError` / `writeMapUsecaseError` | `lib/` never imports `features/`; Zod at API boundary; one gate = one role |
| **Contracts** | `apperrors` codes | `ApiErrorCode` + shared `newPasswordZodString` |

---

## 8. Architectural constraints (DO NOT violate)

1. **No domain → infra imports.** Domain is pure; infra wires it.
2. **No deep feature imports.** Cross-feature uses the slice's `index.ts`.
3. **No business logic in `app/`.** Pages route only; components live in features.
4. **No raw SQL outside `adapter/repository/postgres/sql/`.** Use sqlc.
5. **No env reads outside `internal/config` (BE) or `lib/env.ts` (FE).**
6. **No untyped API responses.** Zod parse at the FE boundary; struct binding + validator on BE.
7. **One file = one component (FE).** Helpers go in `helpers.ts` next to consumers.
8. **One namespace per role.** No mixed URL prefixes (e.g. `/dashboard/*` + `/admin/*`).
9. **No Address module.** Pickup-only fulfillment — no customer delivery/shipping addresses on profile or orders (see §6 Fulfillment model).

---

## 9. Open extensions (foundation present, not wired)

| Hook | Status | Next step |
|------|--------|-----------|
| Server actions | DAL ready | wire mutations from RSC pages |
| Pickup time | `orders.pickup_at` (required at checkout) | still **no** delivery addresses |
| Custom line config | `cart_items.configuration`, `order_items.configuration` jsonb | custom cake templates/inquiries later |

Adding a feature SHOULD follow this spec; deviations require updating this document.
