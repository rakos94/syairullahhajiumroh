package main

import (
	"context"
	"log"
	"time"

	"syairullahhajiumroh/internal/config"
	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

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
	db.Collection("jamaah").Drop(ctx)
	db.Collection("paket").Drop(ctx)
	log.Println("Cleared existing jamaah and paket data")

	paketRepo := repository.NewPaketRepository(db)
	jamaahRepo := repository.NewJamaahRepository(db)

	// Seed paket data
	pakets := []model.Paket{
		{
			Tipe: "haji", Tahun: 2026,
			TanggalKeberangkatan: []model.TanggalKeberangkatanPaket{
				{Nama: "JKG", Tanggal: timePtr(time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))},
				{Nama: "JKS"},
			},
		},
		{Tipe: "umroh", Tahun: 2026, Bulan: 4},
		{Tipe: "umroh", Tahun: 2026, Bulan: 5},
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
	dummies := []model.Jamaah{
		{
			Nama: "ahmad fauzi", NIK: "3201010101010001", NomorPaspor: "A1000001",
			Alamat: "Jl. Merdeka No. 1, Jakarta", NoHP: "081200000001",
			TempatLahir: "Jakarta", TanggalLahir: time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "1234567890", TipeBank: "BRI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
		},
		{
			Nama: "siti aisyah", NIK: "3201010101010002", NomorPaspor: "A1000002",
			Alamat: "Jl. Sudirman No. 5, Bandung", NoHP: "081200000002",
			TempatLahir: "Bandung", TanggalLahir: time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "1234567891", TipeBank: "BNI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: false, KoperSudahDiterima: false,
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
		},
		{
			Nama: "hasan basri", NIK: "3201010101010005", NomorPaspor: "A1000005",
			Alamat: "Jl. Pahlawan No. 8, Medan", NoHP: "081200000005",
			TempatLahir: "Medan", TanggalLahir: time.Date(1970, 5, 12, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "9876543210", TipeBank: "BSI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
		},
		{
			Nama: "nurul hidayah", NIK: "3201010101010006", NomorPaspor: "A1000006",
			Alamat: "Jl. Gatot Subroto No. 3, Makassar", NoHP: "081200000006",
			TempatLahir: "Makassar", TanggalLahir: time.Date(1988, 9, 18, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKS", Tanggal: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "5678901234", TipeBank: "BRI",
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
		},
		{
			Nama: "aisyah putri", NIK: "3201010101010010", NomorPaspor: "A1000010",
			Alamat: "Jl. Veteran No. 9, Denpasar", NoHP: "081200000010",
			TempatLahir: "Denpasar", TanggalLahir: time.Date(1998, 12, 3, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: umrohApril2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Tanggal: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas",
			KoperSudahDiterima: true,
		},
		{
			Nama: "ibrahim malik", NIK: "3201010101010011", NomorPaspor: "A1000011",
			Alamat: "Jl. Kartini No. 22, Banjarmasin", NoHP: "081200000011",
			TempatLahir: "Banjarmasin", TanggalLahir: time.Date(1980, 6, 14, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "laki-laki", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "lunas", NoRekeningHaji: "6677889900", TipeBank: "BSI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: true, KoperSudahDiterima: true,
		},
		{
			Nama: "zainab husna", NIK: "3201010101010012", NomorPaspor: "A1000012",
			Alamat: "Jl. Gajah Mada No. 4, Pontianak", NoHP: "081200000012",
			TempatLahir: "Pontianak", TanggalLahir: time.Date(1993, 10, 25, 0, 0, 0, 0, time.UTC),
			JenisKelamin: "perempuan", PaketID: haji2026,
			TanggalKeberangkatan: &model.KeberangkatanJamaah{Nama: "JKG", Tanggal: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)},
			StatusPembayaran: "dp", NoRekeningHaji: "4455667788", TipeBank: "BRI",
			BatikNasionalSudahDijahit: true, BatikKBIHSudahDiterima: false, KoperSudahDiterima: false,
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
		},
	}

	for i, d := range dummies {
		if err := jamaahRepo.Create(ctx, &d); err != nil {
			log.Printf("[%d] Gagal insert %s: %v", i+1, d.Nama, err)
		} else {
			log.Printf("[%d] Inserted: %s (ID: %s)", i+1, d.Nama, d.ID.Hex())
		}
	}

	log.Println("Seeding selesai!")
}
