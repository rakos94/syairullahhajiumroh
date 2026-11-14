package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username" binding:"required"`
	Password  string             `json:"-" bson:"password"`
	Role      string             `json:"role" bson:"role"`
	// PasswordDefault is true when the password is still the published
	// default; the UI shows a warning. Computed, never stored.
	PasswordDefault bool `json:"password_default,omitempty" bson:"-"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
