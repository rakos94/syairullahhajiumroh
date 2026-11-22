package repository

import (
	"context"
	"time"

	"syairullahhajiumroh/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminRepository struct {
	collection *mongo.Collection
}

func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{
		collection: db.Collection("admins"),
	}
}

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Admin, error) {
	var admin model.Admin
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) Create(ctx context.Context, admin *model.Admin) error {
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, admin)
	if err != nil {
		return err
	}

	admin.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *AdminRepository) FindAll(ctx context.Context) ([]model.Admin, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var admins []model.Admin
	if err := cursor.All(ctx, &admins); err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *AdminRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": update})
	return err
}

func (r *AdminRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// EnsureSuperAdmin creates the super admin if none exists. An existing super
// admin is left untouched (its password is managed in the app), unless
// resetPassword is true, which sets username and password again (recovery).
// It reports whether the account was created or reset.
func (r *AdminRepository) EnsureSuperAdmin(ctx context.Context, username, hashedPassword string, resetPassword bool) (created, reset bool, err error) {
	var existing model.Admin
	err = r.collection.FindOne(ctx, bson.M{"role": "super"}).Decode(&existing)
	switch {
	case err == mongo.ErrNoDocuments:
		now := time.Now()
		_, err = r.collection.InsertOne(ctx, model.Admin{
			Username: username, Password: hashedPassword, Role: "super", CreatedAt: now, UpdatedAt: now,
		})
		return err == nil, false, err
	case err != nil:
		return false, false, err
	case resetPassword:
		_, err = r.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": bson.M{
			"username": username, "password": hashedPassword, "updated_at": time.Now(),
		}})
		return false, err == nil, err
	}
	return false, false, nil
}
