package repository

import (
	"context"
	"strings"
	"time"
	"unicode"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JamaahRepository struct {
	collection *mongo.Collection
}

func NewJamaahRepository(db *mongo.Database) *JamaahRepository {
	return &JamaahRepository{
		collection: db.Collection("jamaah"),
	}
}

// titleCase capitalizes the first letter of each word.
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(strings.ToLower(w))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func (r *JamaahRepository) Create(ctx context.Context, jamaah *model.Jamaah) error {
	jamaah.Nama = titleCase(jamaah.Nama)
	jamaah.CreatedAt = time.Now()
	jamaah.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, jamaah)
	if err != nil {
		return err
	}

	jamaah.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *JamaahRepository) FindAll(ctx context.Context, paketID *primitive.ObjectID, page, limit int) ([]model.Jamaah, int64, error) {
	filter := bson.M{"deleted_at": nil}
	if paketID != nil {
		filter["paket_id"] = *paketID
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []model.Jamaah
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	if results == nil {
		results = []model.Jamaah{}
	}
	return results, total, nil
}

func (r *JamaahRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Jamaah, error) {
	var jamaah model.Jamaah
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&jamaah)
	if err != nil {
		return nil, err
	}
	return &jamaah, nil
}

func (r *JamaahRepository) Update(ctx context.Context, id primitive.ObjectID, jamaah *model.Jamaah) error {
	jamaah.Nama = titleCase(jamaah.Nama)
	jamaah.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deleted_at": nil},
		bson.M{"$set": jamaah},
	)
	return err
}

func (r *JamaahRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now}},
	)
	return err
}

func (r *JamaahRepository) UpdateField(ctx context.Context, id primitive.ObjectID, field string, value interface{}) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deleted_at": nil},
		bson.M{"$set": bson.M{field: value, "updated_at": time.Now()}},
	)
	return err
}
