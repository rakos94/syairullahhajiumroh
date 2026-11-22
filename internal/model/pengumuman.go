package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Pengumuman is an announcement shown on the public home page.
// Isi is plain text (rendered as-is, line breaks preserved).
type Pengumuman struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Judul          string             `json:"judul" bson:"judul" binding:"required"`
	Isi            string             `json:"isi" bson:"isi" binding:"required"`
	Dipublikasikan bool               `json:"dipublikasikan" bson:"dipublikasikan"`
	Disematkan     bool               `json:"disematkan" bson:"disematkan"`
	CreatedBy      string             `json:"created_by,omitempty" bson:"created_by,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}
