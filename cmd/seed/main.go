package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"time"

	"syairullahhajiumroh/internal/config"
	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func timePtr(t time.Time) *time.Time { return &t }

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.MongoDB)

	// Clear existing data
	db.Collection("jamaah").DeleteMany(ctx, bson.M{}) // keep indexes
	db.Collection("paket").DeleteMany(ctx, bson.M{}) // keep indexes
	db.Collection("pengumuman").DeleteMany(ctx, bson.M{}) // keep indexes
	db.Collection("sesi_manasik").DeleteMany(ctx, bson.M{}) // keep indexes
	db.Collection("kehadiran_manasik").DeleteMany(ctx, bson.M{}) // keep indexes
	log.Println("Cleared existing jamaah, paket, pengumuman and manasik data")

	paketRepo := repository.NewPaketRepository(db)
	jamaahRepo := repository.NewJamaahRepository(db)

	// Seed paket data
	pakets := []model.Paket{
		{
			Tipe: "haji", Tahun: 2026, Harga: 45000000,
			TanggalKeberangkatan: []model.TanggalKeberangkatanPaket{
				{Nama: "JKG", Tanggal: timePtr(time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))},
				{Nama: "JKS"},
			},
		},
		{
			Tipe: "umroh", Tahun: 2026, Bulan: 4, Harga: 32500000,
			PilihanHarga: []model.PilihanHarga{
				{Nama: "Quad", Harga: 32500000},
				{Nama: "Triple", Harga: 34500000},
				{Nama: "Double", Harga: 37500000},
			},
		},
		{Tipe: "umroh", Tahun: 2026, Bulan: 5}, // price left unset on purpose (shows "belum diatur")
	}

	paketIDs := make(map[string]primitive.ObjectID) // key: "haji" or "umroh-4" etc
	for i, p := range pakets {
		if err := paketRepo.Create(ctx, &p); err != nil {
			log.Fatalf("Gagal insert paket: %v", err)
		}
		key := p.Tipe
		if p.Bulan > 0 {
			key = p.Tipe + "-" + time.Month(p.Bulan).String()
		}
		paketIDs[key] = p.ID
		p.BuildLabel()
		log.Printf("[Paket %d] Inserted: %s (ID: %s)", i+1, p.Label, p.ID.Hex())
	}

	haji2026 := paketIDs["haji"]
	umrohApril2026 := paketIDs["umroh-April"]
	umrohMei2026 := paketIDs["umroh-May"]

	// Seed jamaah data
	// Manasik rooms: haji jamaah 1,5,11 (laki-laki) -> Gedung A, 2,12,14 (perempuan) -> Gedung B.
	// Nurul Hidayah and Abdullah Rahman are left unassigned on purpose.
	dummies := []model.Jamaah{
		{
			Nama: "ahmad fauzi", NIK: "3201010101010001", NomorPaspor: "A1000001",
			Alamat: "Jl. Merdeka No. 1, Jakarta", NoHP: "081200000001",
			TempatLahir: "Jakarta", TanggalLahir: time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "1234567890", TipeBank: "BRI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
			GedungManasik: "Gedung A", KamarManasik: "101",
			Rombongan: "1", Regu: "1", KetuaRegu: true, KetuaRombongan: true,
		},
		{
			Nama: "siti aisyah", NIK: "3201010101010002", NomorPaspor: "A1000002",
			Alamat: "Jl. Sudirman No. 5, Bandung", NoHP: "081200000002",
			TempatLahir: "Bandung", TanggalLahir: time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "1234567891", TipeBank: "BNI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: false, KoperSudahDiterima: false,
			GedungManasik: "Gedung B", KamarManasik: "201",
			Rombongan: "1", Regu: "1",
			Potongan: 2000000, KeteranganPotongan: "Daftar awal",
		},
		{
			Nama: "muhammad rizki", NIK: "3201010101010003", NomorPaspor: "A1000003",
			Alamat: "Jl. Asia Afrika No. 10, Surabaya", NoHP: "081200000003",
			TempatLahir: "Surabaya", TanggalLahir: time.Date(1978, 1, 5, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: umrohApril2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas",
			KoperSudahDiterima: true,
		},
		{
			Nama: "fatimah zahra", NIK: "3201010101010004", NomorPaspor: "A1000004",
			Alamat: "Jl. Diponegoro No. 20, Semarang", NoHP: "081200000004",
			TempatLahir: "Semarang", TanggalLahir: time.Date(1992, 11, 30, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: umrohApril2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "belum_bayar",
			PilihanHarga: "Triple",
		},
		{
			Nama: "hasan basri", NIK: "3201010101010005", NomorPaspor: "A1000005",
			Alamat: "Jl. Pahlawan No. 8, Medan", NoHP: "081200000005",
			TempatLahir: "Medan", TanggalLahir: time.Date(1970, 5, 12, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "9876543210", TipeBank: "BSI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
			GedungManasik: "Gedung A", KamarManasik: "101",
			Rombongan: "1", Regu: "1",
		},
		{
			Nama: "nurul hidayah", NIK: "3201010101010006", NomorPaspor: "A1000006",
			Alamat: "Jl. Gatot Subroto No. 3, Makassar", NoHP: "081200000006",
			TempatLahir: "Makassar", TanggalLahir: time.Date(1988, 9, 18, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKS", Tanggal: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "5678901234", TipeBank: "BRI",
			Rombongan: "1", Regu: "2",
		},
		{
			Nama: "umar said", NIK: "3201010101010007", NomorPaspor: "A1000007",
			Alamat: "Jl. Ahmad Yani No. 15, Yogyakarta", NoHP: "081200000007",
			TempatLahir: "Yogyakarta", TanggalLahir: time.Date(1982, 2, 28, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: umrohMei2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas",
			KoperSudahDiterima: true,
		},
		{
			Nama: "khadijah aminah", NIK: "3201010101010008", NomorPaspor: "A1000008",
			Alamat: "Jl. Pemuda No. 12, Malang", NoHP: "081200000008",
			TempatLahir: "Malang", TanggalLahir: time.Date(1995, 4, 7, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: umrohMei2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp",
		},
		{
			Nama: "abdullah rahman", NIK: "3201010101010009", NomorPaspor: "A1000009",
			Alamat: "Jl. Imam Bonjol No. 7, Palembang", NoHP: "081200000009",
			TempatLahir: "Palembang", TanggalLahir: time.Date(1975, 8, 20, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKS", Tanggal: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "belum_bayar", NoRekeningHaji: "1122334455", TipeBank: "BNI",
			Rombongan: "1", Regu: "2", KetuaRegu: true, KetuaRombongan: false,
		},
		{
			Nama: "aisyah putri", NIK: "3201010101010010", NomorPaspor: "A1000010",
			Alamat: "Jl. Veteran No. 9, Denpasar", NoHP: "081200000010",
			TempatLahir: "Denpasar", TanggalLahir: time.Date(1998, 12, 3, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: umrohApril2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas",
			KoperSudahDiterima: true,
			PilihanHarga: "Double",
		},
		{
			Nama: "ibrahim malik", NIK: "3201010101010011", NomorPaspor: "A1000011",
			Alamat: "Jl. Kartini No. 22, Banjarmasin", NoHP: "081200000011",
			TempatLahir: "Banjarmasin", TanggalLahir: time.Date(1980, 6, 14, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "6677889900", TipeBank: "BSI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
			GedungManasik: "Gedung A", KamarManasik: "102",
			Rombongan: "1", Regu: "2",
		},
		{
			Nama: "zainab husna", NIK: "3201010101010012", NomorPaspor: "A1000012",
			Alamat: "Jl. Gajah Mada No. 4, Pontianak", NoHP: "081200000012",
			TempatLahir: "Pontianak", TanggalLahir: time.Date(1993, 10, 25, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "4455667788", TipeBank: "BRI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: false, KoperSudahDiterima: false,
			GedungManasik: "Gedung B", KamarManasik: "201",
			Rombongan: "1", Regu: "1",
		},
		{
			Nama: "yusuf hakim", NIK: "3201010101010013", NomorPaspor: "A1000013",
			Alamat: "Jl. Thamrin No. 18, Manado", NoHP: "081200000013",
			TempatLahir: "Manado", TanggalLahir: time.Date(1987, 3, 9, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: umrohMei2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "belum_bayar",
		},
		{
			Nama: "maryam safitri", NIK: "3201010101010014", NomorPaspor: "A1000014",
			Alamat: "Jl. Hayam Wuruk No. 6, Balikpapan", NoHP: "081200000014",
			TempatLahir: "Balikpapan", TanggalLahir: time.Date(1991, 7, 17, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "7788990011", TipeBank: "BNI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
			GedungManasik: "Gedung B", KamarManasik: "201",
			Rombongan: "1", Regu: "2",
		},
	}

	inserted := make([]model.Jamaah, 0, len(dummies))
	for i, d := range dummies {
		if err := jamaahRepo.Create(ctx, &d); err != nil {
			log.Printf("[%d] Gagal insert %s: %v", i+1, d.Nama, err)
		} else {
			log.Printf("[%d] Inserted: %s (ID: %s)", i+1, d.Nama, d.ID.Hex())
			inserted = append(inserted, d)
		}
	}

	// Seed payment receipts and documents (with placeholder image files) so
	// tagihan and the Cek Kamar document checklist have data.
	placeholderPNG, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	writeFile := func(j model.Jamaah, name string) string {
		dir := filepath.Join(cfg.UploadDir, j.ID.Hex())
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Gagal membuat folder upload: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), placeholderPNG, 0644); err != nil {
			log.Fatalf("Gagal menulis file contoh: %v", err)
		}
		return filepath.Join(j.ID.Hex(), name)
	}
	type seedPayment struct {
		dp, pelunasan float64
		docs          []string // doc field names
	}
	payments := map[string]seedPayment{
		"3201010101010001": {dp: 15000000, pelunasan: 30000000, docs: []string{"foto_ktp", "foto_kk", "foto_paspor", "pasfoto"}}, // ahmad: lunas
		"3201010101010002": {dp: 15000000, docs: []string{"foto_ktp", "pasfoto"}},                                                // siti: sisa 30 jt
		"3201010101010005": {dp: 15000000},                                                                                         // hasan: status lunas tapi sisa -> peringatan
		"3201010101010003": {dp: 10000000, pelunasan: 22500000},                                                                     // rizki umroh april: lunas
	}
	for _, j := range inserted {
		p, ok := payments[j.NIK]
		if !ok {
			continue
		}
		if p.dp > 0 {
			jamaahRepo.PushToArray(ctx, j.ID, "bukti_dp", model.BuktiPembayaran{File: writeFile(j, "bukti_dp_seed.png"), Nominal: p.dp})
		}
		if p.pelunasan > 0 {
			jamaahRepo.PushToArray(ctx, j.ID, "bukti_pelunasan", model.BuktiPembayaran{File: writeFile(j, "bukti_pelunasan_seed.png"), Nominal: p.pelunasan})
		}
		for _, field := range p.docs {
			jamaahRepo.UpdateField(ctx, j.ID, field, writeFile(j, field+".png"))
		}
		log.Printf("[Pembayaran] %s: DP %.0f, pelunasan %.0f, %d dokumen", j.Nama, p.dp, p.pelunasan, len(p.docs))
	}

	// Seed sesi manasik: day 1 with some attendance, day 2 empty.
	manasikRepo := repository.NewManasikRepository(db)
	sesi1 := &model.SesiManasik{PaketID: haji2026, Nama: "Hari 1 - Pagi", Tanggal: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)}
	sesi2 := &model.SesiManasik{PaketID: haji2026, Nama: "Hari 2 - Pagi", Tanggal: time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)}
	for _, s := range []*model.SesiManasik{sesi1, sesi2} {
		if err := manasikRepo.CreateSesi(ctx, s); err != nil {
			log.Fatalf("Gagal insert sesi manasik: %v", err)
		}
		log.Printf("[Sesi Manasik] Inserted: %s", s.Nama)
	}
	hadirCount := 0
	for _, j := range inserted {
		if j.PaketID == haji2026 && j.GedungManasik != "" && hadirCount < 4 {
			if _, _, err := manasikRepo.MarkHadir(ctx, &model.KehadiranManasik{SesiID: sesi1.ID, JamaahID: j.ID, DicatatOleh: "admin", Metode: "manual"}); err == nil {
				hadirCount++
			}
		}
	}
	log.Printf("[Sesi Manasik] %d jamaah hadir di %s", hadirCount, sesi1.Nama)

	// Seed pengumuman
	pengumumanRepo := repository.NewPengumumanRepository(db)
	pengumuman := []model.Pengumuman{
		{
			Judul: "Jadwal Manasik Haji 2026 di Asrama Haji",
			Isi: "Manasik haji akan dilaksanakan di Asrama Haji Pondok Gede pada 20-22 Mei 2026.\n\n" +
				"Jamaah diharap hadir pukul 07.00 WIB untuk registrasi dan pembagian kamar. " +
				"Silakan cek gedung dan kamar Anda melalui fitur Cek Kamar di halaman ini.",
			Dipublikasikan: true, Disematkan: true, CreatedBy: "admin",
		},
		{
			Judul: "Pembagian Koper dan Batik KBIH",
			Isi:   "Koper dan batik KBIH dapat diambil di kantor Syairullah setiap Senin-Jumat pukul 09.00-15.00. Harap membawa KTP asli.",
			Dipublikasikan: true, CreatedBy: "admin",
		},
		{
			Judul: "Draft: Info pelunasan BPIH",
			Isi:   "Belum dipublikasikan, menunggu informasi resmi dari Kemenag.",
			CreatedBy: "admin",
		},
	}
	for _, p := range pengumuman {
		if err := pengumumanRepo.Create(ctx, &p); err != nil {
			log.Printf("Gagal insert pengumuman %q: %v", p.Judul, err)
		} else {
			log.Printf("[Pengumuman] Inserted: %s", p.Judul)
		}
	}

	log.Println("Seeding selesai!")
}
