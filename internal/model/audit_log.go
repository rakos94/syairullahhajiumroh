package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FieldChange struct {
	Field    string `json:"field" bson:"field"`
	OldValue string `json:"old_value" bson:"old_value"`
	NewValue string `json:"new_value" bson:"new_value"`
}

type AuditLog struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AdminID       primitive.ObjectID `json:"admin_id" bson:"admin_id"`
	AdminUsername string             `json:"admin_username" bson:"admin_username"`
	EntityType    string             `json:"entity_type" bson:"entity_type"`
	EntityID      primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	EntityLabel   string             `json:"entity_label" bson:"entity_label"`
	Action        string             `json:"action" bson:"action"`
	Description   string             `json:"description" bson:"description"`
	Changes       []FieldChange      `json:"changes,omitempty" bson:"changes,omitempty"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}
