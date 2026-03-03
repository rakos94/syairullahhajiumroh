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

## Development (Frontend Only)

```bash
cd web
npm install
npm run dev
```

Vite dev server runs on port 5173 with API proxy to `localhost:8080`.

## Project Structure

```
cmd/main.go                  # Entry point
internal/
  config/config.go           # Environment configuration
  model/jamaah.go            # Jamaah data model
  repository/                # MongoDB repository
  handler/                   # HTTP handlers (REST API)
  migration/                 # Database migrations
web/                         # React frontend (Vite + Tailwind)
  src/
    api.js                   # API client
    pages/                   # JamaahList, JamaahForm, JamaahDetail
    components/              # Layout, StatusBadge
static.go                    # go:embed for web/dist
docs/                        # Swagger documentation
```

## API Endpoints

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | /api/jamaah | List jamaah (filter: `?paket=haji\|umroh`) |
| POST | /api/jamaah | Tambah jamaah baru |
| GET | /api/jamaah/:id | Detail jamaah |
| PUT | /api/jamaah/:id | Update jamaah |
| DELETE | /api/jamaah/:id | Hapus jamaah |
| POST | /api/jamaah/:id/upload/:docType | Upload dokumen (ktp, kk, paspor, pasfoto, koper_diterima) |
| GET | /api/jamaah/:id/dokumen/:docType | Download dokumen |

## Data Jamaah

- Nama, NIK, Nomor Paspor, Alamat, No HP
- Tanggal Lahir, Jenis Kelamin
- Paket (haji/umroh), Tanggal Keberangkatan
- Status Pembayaran (belum_bayar, dp, lunas)
- No Rekening Haji, Tipe Bank
- Batik Nasional Sudah Dijahit, Batik KBIH Sudah Diterima, Koper Sudah Diterima
- Dokumen: KTP, KK, Paspor, Pas Foto, Foto Koper Diterima
