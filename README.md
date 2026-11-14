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

# Start MongoDB (only) for local development
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

## Docker

Aplikasi (beserta MongoDB) bisa dijalankan sepenuhnya di Docker. Image di-build multi-stage
(frontend → binary Go → Alpine, ±50MB) dan berjalan sebagai user non-root.

```bash
cp .env.example .env      # lalu ubah JWT_SECRET, ADMIN_PASSWORD, MONGO_PASSWORD

make docker-app           # = docker compose up -d --build  → http://localhost:8080
make docker-logs          # lihat log aplikasi
make docker-down          # hentikan semua container
```

- Upload dokumen disimpan di volume `uploads`, database di volume `mongo_data`.
- MongoDB hanya bisa diakses dari mesin itu sendiri (`127.0.0.1:27017`), tidak dari internet.
- Health check: `GET /api/health`.

### Production dengan HTTPS

Scan QR absensi memakai kamera HP, dan browser **hanya mengizinkan kamera lewat HTTPS**.
Untuk server publik, jalankan dengan Caddy (sertifikat Let's Encrypt otomatis):

1. Arahkan DNS domain (misalnya `haji.syairullah.com`) ke IP server.
2. Buka port 80 dan 443 di firewall.
3. Isi `DOMAIN=haji.syairullah.com` di `.env`.
4. Jalankan:

```bash
make docker-https         # = docker compose --profile https up -d --build
```

Setelah itu aplikasi tersedia di `https://<DOMAIN>`. Port 8080 boleh ditutup di firewall.

> **Penting sebelum online:** ganti `JWT_SECRET`, `ADMIN_PASSWORD`, dan `MONGO_PASSWORD`
> dari nilai default. Mengganti `JWT_SECRET` membuat QR absensi lama tidak berlaku
> (jamaah cukup membuka ulang Cek Kamar).

## Seed Data

```bash
# Insert dummy data (3 paket + 14 jamaah, kamar/regu manasik, sesi, pembayaran, pengumuman)
go run ./cmd/seed/
```

> Seed **menghapus** semua data jamaah, paket, pengumuman, dan manasik sebelum mengisi
> data contoh. Hanya untuk development — jangan dijalankan di database production.
> File contoh (bukti bayar, dokumen) ditulis ke `./uploads` lokal, bukan ke volume Docker.

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
