# BOMS — Architecture Spec (Senior 2026+)

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
       Cookie session check                       Redis (sessions, rate limit)
       Cross-feature gates                        Asynq (foundation only)
```

---

## 2. Backend — Hexagonal layout

```
backend/
├── cmd/
│   ├── api/main.go              # HTTP entrypoint + composition root
│   └── genkey/main.go           # Ed25519 key generator CLI
├── db/schema.hcl                # Atlas declarative schema (source of truth)
├── migrations/                  # Atlas versioned SQL (timestamp_*.sql)
├── internal/
│   ├── domain/                  # Entities, value objects, domain errors
│   │   ├── user/                # user, role, audit, errors
│   │   ├── profile/             # customer, staff, admin
│   │   └── session/
│   ├── port/                    # Interfaces (driven + driving)
│   │   ├── user.go, *_profile.go, audit_log.go
│   │   ├── session.go, token.go, password.go, tx.go, health.go
│   ├── usecase/                 # Application services (orchestration)
│   │   ├── auth.go, me.go, admin_user.go, readiness.go
│   ├── service/                 # Domain/application services
│   │   ├── profilesvc/          # Role → profile dispatcher
│   │   └── auditlogger/         # Audit log writer
│   ├── handler/v1/              # HTTP handlers (driving adapters)
│   │   ├── auth.go, me.go, admin_user.go, health.go
│   ├── adapter/
│   │   ├── repository/
│   │   │   ├── postgres/        # sqlc-backed repos + tx context
│   │   │   │   ├── sql/{query,schema}/
│   │   │   │   └── sqlcgen/     # generated
│   │   │   └── redis/           # sessions, client
│   │   └── queue/               # Asynq client (foundation)
│   ├── infrastructure/          # Pure tech: jwt (EdDSA), crypto (argon2id), logger (zap)
│   ├── middleware/              # auth, ratelimit, cors, security_headers, request_meta
│   ├── shared/                  # ctxmeta, errors, response, utils, validator
│   ├── dto/                     # API request/response shapes
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
| Migration | `YYYYMMDDhhmmss_<name>.sql` | `20260524140000_user_profiles.sql` |

---

## 3. Frontend — Feature-Sliced layout

```
frontend/src/
├── proxy.ts                     # Next.js edge proxy (cookie gate + request id)
├── app/                         # Route layer ONLY (thin pages)
│   ├── (public)/                # /, /login, /register
│   ├── (customer)/              # /products, /cart, /orders, /customer/account/*
│   ├── (staff)/                 # /staff/account/*
│   └── (admin)/admin/           # /admin, /admin/users, /admin/account/*
├── features/                    # Feature slices
│   ├── auth/                    # Session, login/register/logout, gates
│   │   ├── api/, schemas/, hooks/, components/, provider/
│   ├── user/                    # /me self-service (canonical user state)
│   │   ├── api/, schemas/, types/, hooks/, components/, lib/
│   └── admin/                   # Operational user management
│       ├── api/, schemas/, types/, hooks/, components/
├── components/ui/               # Primitives (button, input, form, confirm-dialog)
├── lib/
│   ├── api-client.ts, browser-api-client.ts, api-envelope.ts
│   ├── env.ts, utils.ts, validate-next.ts
│   ├── auth/                    # refresh-manager, session helpers
│   ├── dal/                     # Server data access (RSC)
│   ├── errors/                  # api-error, server-api-error
│   └── validation/              # form validation messages
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

### URL conventions (canonical from `constants/routes.ts`)

| Audience | URLs |
|----------|------|
| Public | `/`, `/login`, `/register` |
| Customer | `/products`, `/cart`, `/orders`, `/customer/account/{profile,password,delete}` |
| Staff | `/staff/account/{profile,password}` |
| Baker | `/baker/account/{profile,password}` |
| Manager | `/manager/account/{profile,password}` |
| Admin | `/admin`, `/admin/{products,orders,users}`, `/admin/users/{new,[id]}`, `/admin/account/{profile,password}` |

**Rule:** one role = one namespace. Each role may only access its own URL prefix (enforced by FE `RoleGate` + post-login redirect). Only `admin` is seeded in development (`bootstrap.EnsureDevAdmin`). No mixing of `/dashboard/*` with `/admin/*`. Agent/dev canonical detail: `.cursor/rules/roles.mdc` (local, not committed).

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

- Cross-feature imports → through `@/features/<slice>` barrel only.
- Pages import from `@/features/<slice>` or `@/components/ui/*` only.
- No deep imports like `@/features/admin/components/x` from `app/`.

---

## 4. Security model

| Threat | Control |
|--------|---------|
| Token theft | EdDSA-signed short-lived access JWT + HTTP-only refresh cookie (`/api/v1/auth`-scoped) |
| Session hijack | Redis session store; bearer + cookie hybrid logout; `RequireAuthWithSession` for mutating routes |
| Password attack | Argon2id (params from config); timing-safe dummy hash on login |
| Replay | request-id propagation; refresh rotation |
| CSRF | SameSite=Lax cookie, internal proxy secret on inbound headers, sanitize `X-User-Role`, `X-Request-ID`, `X-Auth-Hint` |
| Bruteforce | Redis-backed rate limit per IP (login/refresh/logout) + per-user admin writes (30/min) |
| RBAC | `RequireRole(Admin)` on `/admin/*`; admin can't modify self; staff self-update only fills `full_name`, `phone` |
| Forced password change | `must_change_password` flag → `RequirePasswordChanged` middleware blocks all routes except `/me` GET and `/me/password` PATCH |
| Audit | All admin mutations write to `audit_logs` with actor/target/before/after |
| Soft delete | `users.deleted_at` (no hard delete from app) |
| Inbound | Strip `x-internal-secret`, `x-user-role`, `x-auth-hint` from client requests in proxy |

---

## 5. Performance principles

- **Backend:** prepared statements via sqlc, paginated queries (max 100), single JOIN for admin user listing, `staleTime` of 5 min on `/me` client-side, `keepPreviousData` for paginated tables.
- **JWT:** stateless access JWT for read routes (`GET /me`); Redis session only on writes — limits Redis QPS.
- **Frontend:** Turbopack build, RSC + Partial Prerendering, route-group code-split per role, Zod parse only at boundary.
- **Polling avoidance:** TanStack Query cache + invalidation on mutation success.
- **Concurrency:** Postgres transactions via `TxManager` for multi-table writes (auth register, role change with profile swap).

---

## 6. Module ownership (canonical sources of truth)

| Concern | Owner |
|---------|-------|
| Session identity (`/me`) | `features/user` (FE) + `usecase/me` (BE) |
| Auth (login/register/logout) | `features/auth` (FE) + `usecase/auth` (BE) |
| Admin user CRUD | `features/admin` (FE) + `usecase/admin_user` (BE) |
| Audit logs | `service/auditlogger` (BE only) |
| Profile entity dispatch | `service/profilesvc` (BE) |
| Routes table | `constants/routes.ts` (FE) |
| Roles enum | `constants/roles.ts` (FE) + `domain/user/role.go` (BE) |
| Validation messages | `lib/validation/` (FE) + `shared/validator/` (BE) |

---

## 7. Verification gates

```bash
# Backend
cd backend
go test ./...           # all green
go build ./...          # all packages compile
make sqlc-check         # generated code in sync (CI)

# Frontend
cd frontend
pnpm typecheck          # tsc --noEmit
pnpm build              # production build
```

CI must run all four. Production deploys block on failure.

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

---

## 9. Open extensions (foundation present, not wired)

| Hook | Status | Next step |
|------|--------|-----------|
| Asynq queue | client only | add `cmd/worker` + task definitions |
| WebSocket | n/a | add `internal/adapter/websocket/` when needed |
| Server actions | DAL ready | wire mutations from RSC pages |

Adding a feature SHOULD follow this spec; deviations require updating this document.
