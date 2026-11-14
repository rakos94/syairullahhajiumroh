package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var bulanIndonesia = [...]string{
	"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

type TanggalKeberangkatanPaket struct {
	Nama    string     `json:"nama" bson:"nama"`
	Tanggal *time.Time `json:"tanggal,omitempty" bson:"tanggal,omitempty"`
}

// PilihanHarga is a price option within a paket, e.g. umroh room type
// "Quad" / "Triple" / "Double".
type PilihanHarga struct {
	Nama  string  `json:"nama" bson:"nama"`
	Harga float64 `json:"harga" bson:"harga"`
}

type Paket struct {
	ID                   primitive.ObjectID          `json:"id" bson:"_id,omitempty"`
	Tipe                 string                      `json:"tipe" bson:"tipe" binding:"required,oneof=haji umroh"`
	Tahun                int                         `json:"tahun" bson:"tahun" binding:"required"`
	Bulan                int                         `json:"bulan,omitempty" bson:"bulan,omitempty"`
	TanggalKeberangkatan []TanggalKeberangkatanPaket `json:"tanggal_keberangkatan,omitempty" bson:"tanggal_keberangkatan,omitempty"`
	// Harga is the package price per jamaah in Rupiah; 0 = not set yet.
	Harga float64 `json:"harga" bson:"harga"`
	// PilihanHarga are optional price options; a jamaah may pick one instead of Harga.
	PilihanHarga []PilihanHarga `json:"pilihan_harga,omitempty" bson:"pilihan_harga,omitempty"`
	Label                string                      `json:"label" bson:"-"`
	CreatedAt            time.Time                   `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at" bson:"updated_at"`
}

// FindPilihanHarga returns the price option with the given name, if any.
func (p *Paket) FindPilihanHarga(nama string) (*PilihanHarga, bool) {
	for i := range p.PilihanHarga {
		if p.PilihanHarga[i].Nama == nama {
			return &p.PilihanHarga[i], true
		}
	}
	return nil, false
}

func (p *Paket) BuildLabel() {
	if p.Tipe == "haji" {
		p.Label = fmt.Sprintf("Haji %d", p.Tahun)
	} else if p.Bulan >= 1 && p.Bulan <= 12 {
		p.Label = fmt.Sprintf("Umroh %s %d", bulanIndonesia[p.Bulan], p.Tahun)
	} else {
		p.Label = fmt.Sprintf("Umroh %d", p.Tahun)
	}
}
