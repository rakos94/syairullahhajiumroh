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

type Paket struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Tipe      string             `json:"tipe" bson:"tipe" binding:"required,oneof=haji umroh"`
	Tahun     int                `json:"tahun" bson:"tahun" binding:"required"`
	Bulan     int                `json:"bulan,omitempty" bson:"bulan,omitempty"`
	Label     string             `json:"label" bson:"-"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
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
