package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KeberangkatanJamaah struct {
	Nama    string    `json:"nama,omitempty" bson:"nama,omitempty"`
	Tanggal time.Time `json:"tanggal" bson:"tanggal"`
}

type BuktiPembayaran struct {
	File    string  `json:"file" bson:"file"`
	Nominal float64 `json:"nominal" bson:"nominal"`
}

// PaketLain points to another registration of the same person (same NIK).
type PaketLain struct {
	JamaahID         primitive.ObjectID `json:"jamaah_id"`
	PaketID          primitive.ObjectID `json:"paket_id"`
	Paket            string             `json:"paket"`
	StatusPembayaran string             `json:"status_pembayaran"`
}

type Jamaah struct {
	ID                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Nama                      string             `json:"nama" bson:"nama" binding:"required"`
	NIK                       string             `json:"nik" bson:"nik" binding:"required"`
	NomorPaspor               string             `json:"nomor_paspor" bson:"nomor_paspor"`
	Alamat                    string             `json:"alamat" bson:"alamat"`
	NoHP                      string             `json:"no_hp" bson:"no_hp"`
	TempatLahir               string             `json:"tempat_lahir" bson:"tempat_lahir"`
	TanggalLahir              time.Time          `json:"tanggal_lahir" bson:"tanggal_lahir"`
	JenisKelamin              string             `json:"jenis_kelamin" bson:"jenis_kelamin" binding:"required,oneof=laki-laki perempuan"`
	PaketID                   primitive.ObjectID `json:"paket_id" bson:"paket_id" binding:"required"`
	Paket                     *Paket             `json:"paket,omitempty" bson:"-"`
	TanggalKeberangkatan      *KeberangkatanJamaah `json:"tanggal_keberangkatan,omitempty" bson:"tanggal_keberangkatan,omitempty"`
	StatusPembayaran          string             `json:"status_pembayaran" bson:"status_pembayaran" binding:"required,oneof=belum_bayar dp lunas"`
	// PilihanHarga is the name of the paket price option chosen (empty = paket price).
	PilihanHarga              string             `json:"pilihan_harga" bson:"pilihan_harga"`
	// Potongan is a per-jamaah discount in Rupiah.
	Potongan                  float64            `json:"potongan" bson:"potongan"`
	KeteranganPotongan        string             `json:"keterangan_potongan" bson:"keterangan_potongan"`
	NoRekeningHaji            string             `json:"no_rekening_haji" bson:"no_rekening_haji"`
	TipeBank                  string             `json:"tipe_bank" bson:"tipe_bank"`
	BatikNasionalSudahDijahit bool               `json:"batik_nasional_sudah_dijahit" bson:"batik_nasional_sudah_dijahit"`
	BatikKBIHSudahDiterima    bool               `json:"batik_kbih_sudah_diterima" bson:"batik_kbih_sudah_diterima"`
	KoperSudahDiterima        bool               `json:"koper_sudah_diterima" bson:"koper_sudah_diterima"`
	// Umroh only (haji uses the two batik fields instead).
	SeragamUmrohSudahDiterima bool `json:"seragam_umroh_sudah_diterima" bson:"seragam_umroh_sudah_diterima"`
	GedungManasik             string             `json:"gedung_manasik" bson:"gedung_manasik"`
	KamarManasik              string             `json:"kamar_manasik" bson:"kamar_manasik"`
	Rombongan                 string             `json:"rombongan" bson:"rombongan"`
	Regu                      string             `json:"regu" bson:"regu"`
	KetuaRegu                 bool               `json:"ketua_regu" bson:"ketua_regu"`
	KetuaRombongan            bool               `json:"ketua_rombongan" bson:"ketua_rombongan"`
	FotoKTP                   string             `json:"foto_ktp" bson:"foto_ktp"`
	FotoKK                    string             `json:"foto_kk" bson:"foto_kk"`
	FotoPaspor                string             `json:"foto_paspor" bson:"foto_paspor"`
	Pasfoto                   string             `json:"pasfoto" bson:"pasfoto"`
	FotoKoperDiterima         string             `json:"foto_koper_diterima" bson:"foto_koper_diterima"`
	BuktiDP                   []BuktiPembayaran  `json:"bukti_dp,omitempty" bson:"bukti_dp,omitempty"`
	BuktiPelunasan            []BuktiPembayaran  `json:"bukti_pelunasan,omitempty" bson:"bukti_pelunasan,omitempty"`
	Keterangan                string             `json:"keterangan" bson:"keterangan"`
	// Computed from paket price and payment receipts; never stored.
	Tagihan *Tagihan `json:"tagihan,omitempty" bson:"-"`
	// Other active registrations of the same NIK (other pakets); only set on
	// the admin detail endpoint.
	PaketLain []PaketLain `json:"paket_lain,omitempty" bson:"-"`
	// Signed access for document links; only set on the admin detail endpoint.
	DokumenAkses *DokumenAkses `json:"dokumen_akses,omitempty" bson:"-"`

	CreatedAt                 time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt                 time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt                 *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
