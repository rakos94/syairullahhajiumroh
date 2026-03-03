# Syairullah Haji & Umroh

Aplikasi manajemen data jamaah haji dan umroh. Backend menggunakan Go (Gin) + MongoDB, frontend menggunakan React + Tailwind CSS yang di-embed ke dalam binary Go.

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

## Setup

```bash
# Clone repository
git clone https://github.com/rakos94/syairullahhajiumroh.git
cd syairullahhajiumroh

# Copy environment file
cp .env.example .env

# Start MongoDB
make docker-up
```

## Running

```bash
# Build frontend + run server
make run
```

- UI: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

## Build

```bash
# Build single binary (includes embedded frontend)
make build

# Run binary
./bin/server
```

## Seed Data

```bash
# Insert dummy data (3 paket + 14 jamaah)
go run ./cmd/seed/
```

## Development (Frontend Only)

```bash
cd web
npm install
npm run dev
```

Vite dev server runs on port 5173 with API proxy to `localhost:8080`.

## Project Structure

```
cmd/
  main.go                      # Entry point
  seed/main.go                 # Seed dummy data
internal/
  config/config.go             # Environment configuration
  model/
    jamaah.go                  # Jamaah data model
    paket.go                   # Paket data model (haji/umroh + tahun/bulan)
  repository/
    jamaah_repository.go       # Jamaah MongoDB repository
    paket_repository.go        # Paket MongoDB repository
  handler/
    jamaah_handler.go          # Jamaah HTTP handlers (REST API)
    paket_handler.go           # Paket HTTP handlers (REST API)
  migration/migration.go       # Database migrations
web/                           # React frontend (Vite + Tailwind)
  src/
    api.js                     # API client
    pages/                     # JamaahList, JamaahForm, JamaahDetail
    components/                # Layout, StatusBadge
static.go                      # go:embed for web/dist
docs/                          # Swagger documentation
```

## API Endpoints

### Paket

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | /api/paket | List paket (filter: `?tipe=haji\|umroh`) |
| POST | /api/paket | Tambah paket baru |
| GET | /api/paket/:id | Detail paket |

### Jamaah

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | /api/jamaah | List jamaah (filter: `?paket_id=...&page=1&limit=10`) |
| POST | /api/jamaah | Tambah jamaah baru |
| GET | /api/jamaah/:id | Detail jamaah |
| PUT | /api/jamaah/:id | Update jamaah |
| DELETE | /api/jamaah/:id | Hapus jamaah (soft delete) |
| POST | /api/jamaah/:id/upload/:docType | Upload dokumen (ktp, kk, paspor, pasfoto, koper_diterima) |
| GET | /api/jamaah/:id/dokumen/:docType | Download dokumen |

## Data Model

### Paket

- Tipe (haji/umroh)
- Tahun (contoh: 2026)
- Bulan (1-12, hanya untuk umroh)
- Label otomatis: "Haji 2026", "Umroh April 2026"

### Jamaah

- Nama (auto title case), NIK, Nomor Paspor, Alamat, No HP
- Tanggal Lahir, Jenis Kelamin
- Paket (referensi ke data paket), Tanggal Keberangkatan
- Status Pembayaran (belum_bayar, dp, lunas)
- No Rekening Haji, Tipe Bank
- Batik Nasional Sudah Dijahit, Batik KBIH Sudah Diterima, Koper Sudah Diterima
- Dokumen: KTP, KK, Paspor, Pas Foto, Foto Koper Diterima

## Fitur

- Server-side pagination pada daftar jamaah
- Soft delete (data tidak dihapus permanen)
- Auto capitalize nama jamaah
- Format tanggal Indonesia (dd/mm/yyyy)
- Preview gambar dokumen dengan lightbox
- Filter jamaah berdasarkan paket
- Single binary deployment (Go embed React build)
