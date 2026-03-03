package repository

import (
	"context"
	"math"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepository struct {
	collection *mongo.Collection
}

func NewAuditLogRepository(db *mongo.Database) *AuditLogRepository {
	return &AuditLogRepository{
		collection: db.Collection("audit_logs"),
	}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	log.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *AuditLogRepository) FindByEntity(ctx context.Context, entityType string, entityID primitive.ObjectID, page, limit int) ([]model.AuditLog, int64, error) {
	filter := bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
	}
	return r.findWithFilter(ctx, filter, page, limit)
}

func (r *AuditLogRepository) FindAll(ctx context.Context, entityType string, page, limit int) ([]model.AuditLog, int64, error) {
	filter := bson.M{}
	if entityType != "" {
		filter["entity_type"] = entityType
	}
	return r.findWithFilter(ctx, filter, page, limit)
}

func (r *AuditLogRepository) findWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]model.AuditLog, int64, error) {
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []model.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	if logs == nil {
		logs = []model.AuditLog{}
	}

	return logs, total, nil
}

func TotalPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}
