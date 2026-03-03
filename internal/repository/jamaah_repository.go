package repository

import (
	"context"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type JamaahRepository struct {
	collection *mongo.Collection
}

func NewJamaahRepository(db *mongo.Database) *JamaahRepository {
	return &JamaahRepository{
		collection: db.Collection("jamaah"),
	}
}

func (r *JamaahRepository) Create(ctx context.Context, jamaah *model.Jamaah) error {
	jamaah.CreatedAt = time.Now()
	jamaah.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, jamaah)
	if err != nil {
		return err
	}

	jamaah.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *JamaahRepository) FindAll(ctx context.Context, paket string) ([]model.Jamaah, error) {
	filter := bson.M{}
	if paket != "" {
		filter["paket"] = paket
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Jamaah
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = []model.Jamaah{}
	}
	return results, nil
}

func (r *JamaahRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Jamaah, error) {
	var jamaah model.Jamaah
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&jamaah)
	if err != nil {
		return nil, err
	}
	return &jamaah, nil
}

func (r *JamaahRepository) Update(ctx context.Context, id primitive.ObjectID, jamaah *model.Jamaah) error {
	jamaah.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": jamaah},
	)
	return err
}

func (r *JamaahRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *JamaahRepository) UpdateField(ctx context.Context, id primitive.ObjectID, field string, value interface{}) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{field: value, "updated_at": time.Now()}},
	)
	return err
}
