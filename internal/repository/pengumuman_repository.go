package repository

import (
	"context"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PengumumanRepository struct {
	collection *mongo.Collection
}

func NewPengumumanRepository(db *mongo.Database) *PengumumanRepository {
	return &PengumumanRepository{collection: db.Collection("pengumuman")}
}

func (r *PengumumanRepository) Create(ctx context.Context, p *model.Pengumuman) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	res, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// FindAll returns announcements, pinned first then newest first.
// When onlyPublished is true, drafts are excluded (used by the public home page).
func (r *PengumumanRepository) FindAll(ctx context.Context, onlyPublished bool, limit int64) ([]model.Pengumuman, error) {
	filter := bson.M{}
	if onlyPublished {
		filter["dipublikasikan"] = true
	}
	opts := options.Find().SetSort(bson.D{{Key: "disematkan", Value: -1}, {Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	results := []model.Pengumuman{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *PengumumanRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Pengumuman, error) {
	var p model.Pengumuman
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PengumumanRepository) Update(ctx context.Context, id primitive.ObjectID, p *model.Pengumuman) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"judul":          p.Judul,
		"isi":            p.Isi,
		"dipublikasikan": p.Dipublikasikan,
		"disematkan":     p.Disematkan,
		"updated_at":     time.Now(),
	}})
	return err
}

func (r *PengumumanRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
