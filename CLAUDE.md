# CLAUDE.md

Haji/umroh jamaah management. Go 1.25 (Gin) + MongoDB backend; React 19 + Vite + Tailwind 4 SPA embedded into the Go binary (`static.go` → `web/dist`). UI text, domain names, error messages are **Bahasa Indonesia**; code comments English.

## Commands
- `make docker-up` Mongo only → `make run` (builds web, runs `cmd/main.go`, :8080)
- `cd web && npm run dev` (:5173, proxies `/api` → :8080) · `npm run lint`
- `go build ./... && go vet ./...` — no Go tests exist
- `go run ./cmd/seed/` — WIPES jamaah/paket/pengumuman/manasik, dev only
- Docker: `make docker-app` / `docker-https` (Caddy) / `backup` / `restore BACKUP=...`
- Swagger: swaggo annotations; `docs/` is generated (`swag init -g cmd/main.go`)

## Layout (backend: handler → repository → Mongo; no service layer)
- `cmd/main.go` — wiring, all route groups, startup security checks, SPA fallback
- `internal/config` — env (`.env.example`); `Production` = `GIN_MODE=release`
- `internal/model` — structs + pure domain logic (`HitungTagihan`, `StatusOtomatis` in `tagihan.go`)
- `internal/repository` — one per collection
- `internal/handler` — `RegisterRoutes(protected)` / `RegisterPublicRoutes(api)` per handler
  - `audit.go`: `logAudit` / `logAuditWithChanges` + `diffField` — call on every mutation
  - `status_pembayaran.go`: payment status derived from receipts when paket has price (not manually editable)
  - `document_access.go`: HMAC-signed doc links (2h TTL); `/jamaah/:id/dokumen/:docType` is outside auth
  - `manasik_token.go`: QR token `SYH1.<id>.<hmac>`; `portal_handler.go`: public Cek Kamar (rate limited)
- `internal/middleware` — JWT `AuthRequired` (sets `admin_id`, `admin_username`), rate limiter, security headers
- `internal/migration/migration.go` — ordered list `001_…`–`009_…`, recorded in `migrations` collection; append new ones, never edit applied ones

## Frontend (`web/src`)
- `api.js` — all fetch calls via `authFetch` (Bearer from localStorage, 401 → /login); images compressed client-side before upload
- `App.jsx` routes; public: `/`, `/login`, `/cek-kamar`; rest behind `ProtectedRoute` + `Layout`
- `pages/` one file per screen; shared bits in `components/` (`RupiahInput`, `NominalDialog`, `CopyButton`, `StatusBadge`, `ManasikQr`)
- `utils/format.js` — `formatRupiah`, `terbilang`; dates shown dd/mm/yyyy

## Domain glossary
jamaah=pilgrim · paket=package (haji/umroh, tahun, bulan, prices) · tagihan=billing · DP/lunas/belum_bayar=payment status · bukti=receipt · manasik=rituals training (sesi=session, kehadiran=attendance, kamar/regu=room/group) · pengumuman=announcement · riwayat=audit log · NIK=national ID

## Conventions
- Soft delete (`deleted_at`) for jamaah; NIK unique
- Roles: `"super"` (created from env on first boot only; manages admins) and `"admin"`
- Prod refuses to start with default/short `JWT_SECRET`; never weaken these checks
- When adding routes/fields/features, update README "API Endpoints" / "Data Model" / "Fitur"
