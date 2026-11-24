# Syairullah Haji & Umroh

Aplikasi manajemen data jamaah haji dan umroh. Backend menggunakan Go (Gin) + MongoDB, frontend menggunakan React + Tailwind CSS yang di-embed ke dalam binary Go.

## Prerequisites

- Go 1.25+
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

## Keamanan

Sebelum aplikasi bisa diakses dari internet:

1. **`JWT_SECRET`** — isi dengan nilai acak: `openssl rand -hex 32`. Di production (Docker)
   aplikasi **menolak jalan** jika masih default atau kurang dari 32 karakter, karena dengan
   secret default siapa pun bisa membuat token admin palsu. Mengganti secret membuat semua
   sesi login dan QR absensi lama tidak berlaku (jamaah cukup membuka ulang Cek Kamar).
2. **Password super admin** — `ADMIN_USERNAME`/`ADMIN_PASSWORD` hanya dipakai saat super
   admin pertama kali dibuat; setelah itu password **diganti di menu Admin** dan tidak lagi
   ditimpa dari `.env`. Selama password masih `admin123`, admin melihat banner merah.
   - Lupa password: set `ADMIN_RESET_PASSWORD=true` dan `ADMIN_PASSWORD=<baru>` di `.env`,
     jalankan `make docker-app`, login, lalu kembalikan `ADMIN_RESET_PASSWORD=false`.
3. **Password MongoDB** — `MONGO_PASSWORD` hanya berlaku saat volume database pertama kali
   dibuat. Untuk mengganti di instalasi yang sudah berjalan:
   ```bash
   docker compose exec mongodb mongosh -u admin -p <password-lama> --authenticationDatabase admin \
     --eval 'db.getSiblingDB("admin").changeUserPassword("admin", "<password-baru>")'
   # lalu ubah MONGO_PASSWORD (dan MONGODB_URI) di .env, kemudian:
   make docker-app
   ```

Yang sudah diatur otomatis:

- **Dokumen jamaah** (KTP, KK, paspor, bukti bayar) hanya bisa dibuka admin yang login.
  Halaman detail memakai link bertanda tangan yang kedaluwarsa setelah 2 jam.
- **Rate limit** login dan Cek Kamar tidak bisa diakali dengan header `X-Forwarded-For`
  palsu; header itu hanya dipercaya dari Caddy (IP tetap `172.30.57.10`).
- Dengan `make docker-https`, port aplikasi hanya terbuka di `127.0.0.1`, jadi semua akses
  lewat HTTPS. Caddy menambahkan HSTS; aplikasi menambahkan header keamanan dasar.
- Swagger (`/swagger`) mati di production kecuali `ENABLE_SWAGGER=true`.

## Catatan Upgrade

**Status pembayaran otomatis.** Untuk paket yang sudah punya harga, status pembayaran kini
dihitung dari bukti DP & pelunasan yang diunggah (Belum Bayar / DP / Lunas) dan tidak bisa
diubah manual. Saat pertama kali dijalankan, migrasi `009` menyesuaikan status semua jamaah
dengan bukti yang ada.

- **Sebelum upgrade:** buka menu **Tagihan** dan periksa jamaah dengan peringatan
  "Status Lunas tetapi masih ada sisa tagihan". Unggah bukti pembayaran yang belum ada
  (misalnya pembayaran tunai), karena tanpa bukti status mereka akan berubah menjadi DP atau
  Belum Bayar.
- **Setelah upgrade:** setiap perubahan dari migrasi tercatat di **Riwayat** dengan
  pengguna `sistem` beserta status lama dan barunya.

## Backup & Restore

`make docker-app` juga menjalankan service **backup** yang setiap hari pukul `BACKUP_TIME`
(default 02:00 WIB) menyimpan database + semua dokumen upload ke `./backups/backup-<tanggal>/`.
Hanya **`BACKUP_KEEP`** backup terbaru yang disimpan (default 7); yang lebih lama dihapus otomatis.

```bash
make backup                                   # backup sekarang
make backup-list                              # daftar backup
make restore BACKUP=backup-20261114-020000    # restore (data sekarang DIGANTI, minta konfirmasi)
```

> Backup tersimpan di server yang sama. Untuk perlindungan dari kerusakan server, salin folder
> `backups/` secara berkala ke tempat lain (misalnya dengan `rclone` ke Google Drive/S3).

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
  main.go                  # Entry point: wiring, routes, cek keamanan saat start
  seed/main.go             # Seed data contoh (development)
internal/
  config/                  # Konfigurasi dari environment (.env)
  model/                   # Struct data + logika domain (perhitungan tagihan, status otomatis)
  repository/              # Akses MongoDB (satu file per koleksi)
  handler/                 # HTTP handler (REST API), audit log, link dokumen bertanda tangan, token QR manasik
  middleware/              # JWT auth, rate limit, security headers
  migration/migration.go   # Migrasi database berurutan (001–009)
web/src/                   # React frontend (Vite + Tailwind)
  api.js                   # Semua pemanggilan API
  pages/                   # Satu file per halaman
  components/              # Komponen bersama (Layout, RupiahInput, NominalDialog, ManasikQr, ...)
  utils/format.js          # formatRupiah, terbilang
scripts/backup.sh          # Backup/restore MongoDB + upload
static.go                  # go:embed web/dist
docs/                      # Swagger (generated: swag init -g cmd/main.go)
```

## API Endpoints

Semua endpoint di bawah prefix `/api`. Kecuali ditandai **publik**, endpoint butuh header
`Authorization: Bearer <token>`. Detail lengkap di Swagger.

| Area | Endpoint |
|------|----------|
| Health (publik) | `GET /health` |
| Auth | `POST /auth/login` (publik, rate limit), `GET /auth/me` |
| Admin (super admin) | `GET/POST /admin`, `PUT/DELETE /admin/:id` |
| Paket | `GET/POST /paket` (`?tipe=haji\|umroh`), `GET/PUT/DELETE /paket/:id` |
| Jamaah | `GET/POST /jamaah` (`?paket_id&search&status_pembayaran&kelengkapan&page&limit`), `GET/PUT/DELETE /jamaah/:id`, `PUT /jamaah/bulk-manasik` |
| Dokumen | `POST /jamaah/:id/upload/:docType`, `DELETE /jamaah/:id/dokumen/:docType`, `GET /jamaah/:id/dokumen/:docType` (admin atau link bertanda tangan) |
| Tagihan | `GET /tagihan`, `GET /tagihan/export` (`?paket_id&belum_lunas=true`) |
| Manasik | `GET /manasik/jamaah`, `GET /manasik/export`, `POST /manasik/import`, `GET/POST /manasik/sesi`, `GET/DELETE /manasik/sesi/:id`, `POST /manasik/sesi/:id/hadir`, `DELETE /manasik/sesi/:id/hadir/:jamaahId` |
| Pengumuman | `GET/POST /pengumuman`, `PUT/DELETE /pengumuman/:id`; `GET /portal/pengumuman` (publik) |
| Portal jamaah (publik) | `POST /portal/cek-kamar` (rate limit) |
| Lainnya | `GET /statistics`, `GET /audit-logs`, `GET /audit-logs/:entityType/:entityId` |

`docType`: `foto_ktp`, `foto_kk`, `foto_paspor`, `pasfoto`, `foto_koper_diterima`, `bukti_dp`, `bukti_pelunasan`
(bukti bisa lebih dari satu file, dipilih dengan `?index=`).

## Data Model

### Paket

- Tipe (haji/umroh), Tahun, Bulan (hanya umroh) → label otomatis: "Haji 2026", "Umroh April 2026"
- Tanggal keberangkatan (bisa lebih dari satu)
- Harga, atau beberapa pilihan harga (nama + harga)

### Jamaah

- Identitas: Nama (auto title case), NIK, Nomor Paspor, Alamat, No HP, Tempat/Tanggal Lahir, Jenis Kelamin
- Paket, Tanggal Keberangkatan, Pilihan Harga, Potongan (+ keterangan)
- Pembayaran: Status (belum_bayar / dp / lunas — otomatis dari bukti jika paket punya harga), Bukti DP & Pelunasan (file + nominal), No Rekening Haji, Tipe Bank
- Perlengkapan: haji → Batik Nasional, Batik KBIH, Koper; umroh → Seragam Umroh, Koper
- Manasik: Gedung, Kamar, Rombongan, Regu, Ketua Regu/Rombongan
- Dokumen: KTP, KK, Paspor, Pas Foto, Foto Koper Diterima
- Satu orang bisa ikut beberapa paket (mis. umroh lalu haji): tiap paket satu record jamaah. NIK dan nomor paspor unik **per paket** (record yang sudah dihapus tidak dihitung). Data diri (nama, NIK, paspor, alamat, no HP, tempat/tanggal lahir, jenis kelamin) disinkronkan otomatis ke semua record dengan NIK yang sama saat diedit atau saat didaftarkan ke paket baru (tanggal lahir harus sama dengan pendaftaran sebelumnya, untuk mencegah salah ketik NIK), dan dokumen identitas (KTP, KK, paspor, pas foto) dipakai bersama: sekali unggah berlaku di semua paket. Tagihan, bukti pembayaran, foto koper, dan manasik tetap per record.

### Lainnya

- **Admin** — role `super` (kelola admin) atau `admin`
- **Sesi Manasik** + kehadiran (scan QR jamaah)
- **Pengumuman** — tampil di halaman publik
- **Audit Log** — riwayat semua perubahan (siapa, kapan, field lama/baru)

## Fitur

- Login admin (JWT) dengan manajemen admin oleh super admin
- Daftar jamaah dengan pagination server-side, pencarian, dan filter (paket, status, kelengkapan dokumen)
- Soft delete (data tidak dihapus permanen)
- Upload dokumen dengan kompresi gambar di browser, preview lightbox, link dokumen kedaluwarsa
- Tagihan per jamaah (harga, potongan, DP, pelunasan, sisa) dengan export
- Manasik: pembagian kamar/regu (import/export), sesi, absensi via scan QR
- Portal jamaah publik: Cek Kamar (menampilkan semua paket yang diikuti NIK tersebut) + QR absensi, pengumuman
- "Daftarkan ke Paket Lain" dari detail jamaah: data diri disalin ke pendaftaran paket baru
- Dashboard statistik dan riwayat perubahan (audit log)
- Format Rupiah + terbilang, tanggal Indonesia (dd/mm/yyyy), tombol salin (NIK, dll.)
- Single binary deployment (Go embed React build), backup harian otomatis
