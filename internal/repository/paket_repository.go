package repository

import (
	"context"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaketRepository struct {
	collection *mongo.Collection
}

func NewPaketRepository(db *mongo.Database) *PaketRepository {
	return &PaketRepository{
		collection: db.Collection("paket"),
	}
}

func (r *PaketRepository) Create(ctx context.Context, paket *model.Paket) error {
	paket.CreatedAt = time.Now()
	paket.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, paket)
	if err != nil {
		return err
	}

	paket.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PaketRepository) FindAll(ctx context.Context, tipe string) ([]model.Paket, error) {
	filter := bson.M{}
	if tipe != "" {
		filter["tipe"] = tipe
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Paket
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = []model.Paket{}
	}

	for i := range results {
		results[i].BuildLabel()
	}
	return results, nil
}

func (r *PaketRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Paket, error) {
	var paket model.Paket
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&paket)
	if err != nil {
		return nil, err
	}
	paket.BuildLabel()
	return &paket, nil
}

func (r *PaketRepository) Update(ctx context.Context, id primitive.ObjectID, paket *model.Paket) error {
	paket.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"tipe":       paket.Tipe,
			"tahun":      paket.Tahun,
			"bulan":      paket.Bulan,
			"updated_at": paket.UpdatedAt,
		}},
	)
	return err
}

func (r *PaketRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *PaketRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.Paket, error) {
	if len(ids) == 0 {
		return []model.Paket{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Paket
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for i := range results {
		results[i].BuildLabel()
	}
	return results, nil
}
