package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SesiManasik is one manasik session (e.g. "Hari 1 - Pagi") for a haji paket.
type SesiManasik struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PaketID   primitive.ObjectID `json:"paket_id" bson:"paket_id" binding:"required"`
	Nama      string             `json:"nama" bson:"nama" binding:"required"`
	Tanggal   time.Time          `json:"tanggal" bson:"tanggal"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`

	// Computed, not stored.
	JumlahHadir int64 `json:"jumlah_hadir" bson:"-"`
}

// KehadiranManasik records that a jamaah attended a session.
// (sesi_id, jamaah_id) is unique.
type KehadiranManasik struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SesiID      primitive.ObjectID `json:"sesi_id" bson:"sesi_id"`
	JamaahID    primitive.ObjectID `json:"jamaah_id" bson:"jamaah_id"`
	Waktu       time.Time          `json:"waktu" bson:"waktu"`
	DicatatOleh string             `json:"dicatat_oleh" bson:"dicatat_oleh"`
	Metode      string             `json:"metode" bson:"metode"` // "qr" | "manual"
}
