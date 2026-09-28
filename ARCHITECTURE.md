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
       Cross-feature gates                        Asynq (foundation only)
             │                                            │
       Browser ── ws(s)://…/ws?ticket= ──▶  Realtime listener (push-only, own port)
```

---

## 2. Backend — Hexagonal layout

```
backend/
├── cmd/
│   ├── api/main.go              # HTTP entrypoint + composition root
│   ├── worker/main.go           # Event worker: sweeps undelivered outbox events, prunes old ones
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
│   │   └── session/
│   ├── port/                    # Interfaces (driven + driving)
│   │   ├── user.go, *_profile.go, audit_log.go
│   │   ├── outbox.go            # EventOutbox + EventPublisher
│   │   ├── session.go, token.go, password.go, tx.go, health.go
│   ├── usecase/                 # Application services (orchestration)
│   │   ├── auth.go, me.go, admin_user.go, manager_category.go, manager_product.go, catalog.go, readiness.go
│   ├── service/                 # Domain/application services
│   │   ├── profilesvc/          # Role → profile dispatcher
│   │   ├── auditlogger/         # Audit log writer
│   │   └── eventdispatch/       # Delivers outbox events after commit; sweep + prune for the worker
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
│   │   └── queue/               # Asynq client (foundation)
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
│   ├── (public)/                # /, /login, /register (+ PublicSessionGate layout)
│   ├── (customer)/              # /products, /cart, /orders, /customer/account/*
│   ├── (staff)/                 # /staff/orders, /staff/account/*
│   ├── (baker)/                 # /baker/account/*
│   ├── (manager)/               # /manager, /manager/categories, /manager/products, /combos, /discount-codes, /account/*
│   └── (admin)/admin/           # /admin, /admin/users, /admin/account/*
├── features/                    # Feature slices (auth | user | admin | manager | staff | baker | customer | catalog)
│   ├── auth/                    # api/, schemas/, hooks/, components/, lib/, provider/
│   ├── user/                    # api/, schemas/, types/, hooks/, components/
│   ├── admin/                   # api/, schemas/, types/, hooks/, components/
│   ├── manager/                 # catalog CRUD → /api/v1/manager/*
│   ├── staff/                   # order queue → /api/v1/staff/orders/*
│   ├── baker/                   # production queue → /api/v1/baker/production/*
│   ├── customer/                # cart, checkout, orders → /api/v1/cart/*, /orders/*
│   └── catalog/                 # storefront browse → /api/v1/catalog/*
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
| Identity API | `features/user` owns `GET /me` only |

**Rules:** no `features/index.ts` meta-barrel; `app/` pages stay thin; RBAC gates are UX — backend enforces roles.

### URL conventions (canonical from `constants/routes.ts`)

| Audience | URLs |
|----------|------|
| Public | `/`, `/login`, `/register` |
| Customer | `/products`, `/cart`, `/orders`, `/customer/account/{profile,password,delete}` |
| Staff | `/staff/orders`, `/staff/orders/{id}`, `/staff/account/{profile,password}` |
| Baker | `/baker/production`, `/baker/production/:id`, `/baker/account/{profile,password}` |
| Manager | `/manager`, `/manager/categories`, `/manager/products`, `/manager/account/{profile,password}` |
| Admin | `/admin`, `/admin/users`, `/admin/users/{new,[id]}`, `/admin/account/profile` (profile + password) |

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
| Session hijack | Redis session store; bearer + cookie hybrid logout; `RequireAuthWithSession` for mutating routes |
| Password attack | Argon2id (params from config); timing-safe dummy hash on login |
| Replay | request-id propagation; refresh rotation |
| CSRF | SameSite=Lax cookie, internal proxy secret on inbound headers, sanitize `X-User-Role`, `X-Request-ID`, `X-Auth-Hint` |
| Bruteforce | Redis-backed rate limit per IP (login/refresh/logout) + per-user admin writes (30/min), manager catalog writes (30/min, `RATE_LIMIT_REDIS_MANAGER_WRITE_*`), order mutations — checkout + staff/baker status (20/min, `RATE_LIMIT_REDIS_ORDER_WRITE_*`) + self-service account writes — `PATCH /me`, `PATCH /me/password`, `DELETE /me` (10/min, `RATE_LIMIT_REDIS_SELF_WRITE_*`) + discount code attempts per customer — `PUT /cart/discount` (10/15 min, `RATE_LIMIT_REDIS_DISCOUNT_ATTEMPT_*`) + manager Cloudinary signatures (20/min, `RATE_LIMIT_REDIS_MANAGER_MEDIA_*`) + realtime tickets per user (60/min, `RATE_LIMIT_REDIS_REALTIME_TICKET_*`) |
| RBAC | `RequireRole(Admin)` on `/admin/*`; admin can't modify self; staff self-update only fills `full_name`, `phone` |
| Forced password change | `must_change_password` flag → `RequirePasswordChanged` middleware blocks all routes except `/me` GET and `/me/password` PATCH |
| Audit | All admin mutations write to `audit_logs` with actor/target/before/after |
| Soft delete | `users.deleted_at` (no hard delete from app) |
| Realtime socket | The only public surface besides `/health` and `/ready`, on its own port and path (`/ws`). Admission refuses a missing or unlisted `Origin`, a malformed token and anything over the per-address or per-process redeem rate before touching Redis, then needs a single-use ticket issued through the BFF to a signed-in session (32 random bytes, stored hashed, 30 s, redeemed with `GETDEL`) whose session still exists, and a free slot under the per-user and per-process caps. Handshakes are bounded by `REALTIME_HANDSHAKE_TIMEOUT`. Channels are chosen from the ticket, never from the browser; the socket carries change notices only and refuses any data frame from the browser |
| Multi-tab sessions | Tabs share one refresh cookie and the API revokes every session when a refresh token is spent twice, so tabs refresh one at a time (Web Locks) and pass the new access token to the others (`BroadcastChannel`, same origin, adopted only by tabs of the same account) |
| Inbound | Strip `x-internal-secret`, `x-user-role`, `x-auth-hint` and the forwarding headers (`x-forwarded-for`, `x-real-ip`, `forwarded`, `x-client-ip`) from client requests in proxy |
| Landing flash | The access token is memory-only, so a returning visitor is anonymous until the session is restored. Fiber issues a role cookie beside the session cookie (`COOKIE_ROLE_NAME`) and `proxy.ts` lands them at the edge (`lib/routing/signed-in-landing.ts`). A **navigation hint only** — never authorization: every byte of data stays behind the API session check, and a forged or stale role reaches a page whose gate sends it back. **Document navigations only** (`Sec-Fetch-Dest: document`, never `/api/*`): on a client navigation the hint would fight the gate that knows the real session, and the two would bounce a visitor between them — including away from the sign-in page, exactly when the session cannot be restored. The redirect carries `Cache-Control: no-store, private` and `Vary: Cookie`, because a cookie decided it |
| Visitor identity | The API only ever sees the BFF. `proxy.ts` resolves the visitor from what the hosting edge reported (`resolveClientIp`) before it drops the forwarding headers, stamps the checked address as `X-Client-IP`, and the BFF forwards that stamp (`stampedClientIp`, `lib/server/client-ip.ts`); `middleware.ClientIP` reads it, falling back to the socket address when it is absent or malformed. Rate-limit buckets and `audit_logs.ip` would otherwise all be the BFF |
| Customer data by role | Each role gets only the customer data its work needs, decided in the usecase DTO mappers: staff get name and email on the order list and the phone as well on one order (`StaffOrderCustomerResponse`) to coordinate its pickup — a list page never hands out every customer's number; bakers get the display name only (`BakerOrderCustomerResponse`) — the kitchen never contacts a customer |
| Abandoned request | `HTTP_REQUEST_TIMEOUT` deadlines the handler context (below `HTTP_WRITE_TIMEOUT`), `POSTGRES_STATEMENT_TIMEOUT` caps one statement — a caller that walks away cannot hold database connections |

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
| Auth (login/register/logout) | `features/auth` (FE) + `usecase/auth` (BE) |
| Admin user CRUD | `features/admin` (FE) + `usecase/admin_user` (BE) |
| Manager catalog CRUD | `features/manager` (FE) + `usecase/manager_category` + `usecase/manager_product` + `manager_combo` + `manager_discount_code` (BE) |
| Product images (Cloudinary) | `lib/cloudinary/*` + `components/ui/catalog-image-list-field` (FE) + `usecase/manager_media` + `service/cloudinary` (BE) |
| Storefront catalog browse | `features/catalog` (FE) + `usecase/catalog` (BE) — API path `/catalog/*` |
| Customer cart & checkout | `features/customer` (FE) + `usecase/cart` + `usecase/order` (BE) — `/cart/*`, `/orders/*` (session + server pricing; **pickup-only**, see below) |
| Staff order queue | `features/staff` (FE) + `usecase/staff_order` (BE) — `/staff/orders/*` (list, detail, status transitions) |
| Baker production queue | `features/baker` (FE) + `usecase/baker_order` (BE) — `/baker/production/*` (list, detail, kitchen transitions) |
| Audit logs | `service/auditlogger` (BE only) |
| Event delivery | `domain/event` + `port/outbox.go` + `adapter/repository/postgres/outbox_repository.go` + `adapter/eventbus` + `service/eventdispatch` + `cmd/worker` (BE only) |
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
| Recover | `cmd/worker` sweeps rows still unpublished `OUTBOX_SWEEP_GRACE` (longer than twice the dispatch timeout) after they were written — at start and then every `OUTBOX_SWEEP_INTERVAL`, draining full batches at once. `FOR UPDATE SKIP LOCKED` means two workers never hold the same row at once; a sweep that published but could not mark rolls back and sends again. The worker validates only the database, Redis and outbox settings (`config.LoadWorker`) |
| Retain | published rows are deleted after `OUTBOX_RETENTION` — delivery records, not business data, so no soft delete |
| Clock | `created_at` defaults to `clock_timestamp()` and every age (sweep grace, retention) is measured in SQL, so the API and the worker never compare two hosts' clocks |

Delivery is **at least once** and **unordered across transactions** (each commit is delivered on its own): an event may arrive twice or after a later one, and subscribers use it only as a hint to refetch through the API. The bus is Redis Pub/Sub (`adapter/eventbus`): channel `boms:events:user:{id}` per user and `boms:events:role:{role}` per role, message `{id, type, at, data}` — identifiers and labels, never the changed record.

| Topic | Written by | Audience | Data |
|-------|------------|----------|------|
| `order.created` | checkout | the customer · staff | `order_id`, `status` |
| `order.status_changed` | staff and baker status moves | the customer · staff · baker only when the order enters, leaves or moves within the statuses bakers see (`Status.VisibleToBaker`) | `order_id`, `status` |

Topics and their audiences live with the aggregate that raises them (`domain/order/event.go`), as audit actions do.

The API process delivers after commit; run `cmd/worker` beside it (`make run-worker`) to recover what that delivery missed.

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
| Channels | every ticket hears its own `boms:events:user:{id}`; staff and baker also hear their role channel. A customer never shares a channel |
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
| **Checkout / orders** | `orders`: pricing, discount snapshot, `status`, required `pickup_at` at checkout (8:00–18:00 bakery local, 2h lead, 14d max), line items with `configuration` jsonb — **no** shipping/delivery address fields. |
| **Storefront `BRAND.addressLine`** | Static marketing copy for footer “Visit us” (`constants/brand.ts`) — the **bakery location**, not per-customer data. |
| **Marketing copy** | UI may say “pickup” but must not imply saved delivery addresses or ship-to-door unless a feature is implemented. |

**Order status (schema v2):** `pending` → `confirmed` → `in_production` → `ready` → `fulfilled` | `cancelled`.

| Role | Allowed transitions |
|------|---------------------|
| **Staff** | `pending`→`confirmed`\|`cancelled`; `confirmed`\|`in_production`→`cancelled`; `ready`→`fulfilled`\|`cancelled` |
| **Baker** | `confirmed`→`in_production`; `in_production`→`ready` — `GET/PATCH /api/v1/baker/production/*` |

**Staff workflow:** counter confirm/cancel and handoff when `ready`; kitchen progress is baker-owned.

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
| Asynq queue | client only — `cmd/worker` runs the outbox sweeper, not Asynq tasks | add task handlers to the worker with the first background job |
| Server actions | DAL ready | wire mutations from RSC pages |
| Pickup time | `orders.pickup_at` (required at checkout) | still **no** delivery addresses |
| Custom line config | `cart_items.configuration`, `order_items.configuration` jsonb | custom cake templates/inquiries later |

Adding a feature SHOULD follow this spec; deviations require updating this document.
