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

type ManasikRepository struct {
	sesi      *mongo.Collection
	kehadiran *mongo.Collection
}

func NewManasikRepository(db *mongo.Database) *ManasikRepository {
	return &ManasikRepository{
		sesi:      db.Collection("sesi_manasik"),
		kehadiran: db.Collection("kehadiran_manasik"),
	}
}

func (r *ManasikRepository) CreateSesi(ctx context.Context, s *model.SesiManasik) error {
	s.CreatedAt = time.Now()
	res, err := r.sesi.InsertOne(ctx, s)
	if err != nil {
		return err
	}
	s.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ManasikRepository) FindSesiByPaket(ctx context.Context, paketID primitive.ObjectID) ([]model.SesiManasik, error) {
	opts := options.Find().SetSort(bson.D{{Key: "tanggal", Value: 1}, {Key: "created_at", Value: 1}})
	cursor, err := r.sesi.Find(ctx, bson.M{"paket_id": paketID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	list := []model.SesiManasik{}
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	for i := range list {
		n, err := r.kehadiran.CountDocuments(ctx, bson.M{"sesi_id": list[i].ID})
		if err != nil {
			return nil, err
		}
		list[i].JumlahHadir = n
	}
	return list, nil
}

func (r *ManasikRepository) FindSesiByID(ctx context.Context, id primitive.ObjectID) (*model.SesiManasik, error) {
	var s model.SesiManasik
	if err := r.sesi.FindOne(ctx, bson.M{"_id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSesi removes the session and all of its attendance records.
func (r *ManasikRepository) DeleteSesi(ctx context.Context, id primitive.ObjectID) error {
	if _, err := r.kehadiran.DeleteMany(ctx, bson.M{"sesi_id": id}); err != nil {
		return err
	}
	_, err := r.sesi.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ManasikRepository) FindKehadiran(ctx context.Context, sesiID primitive.ObjectID) ([]model.KehadiranManasik, error) {
	opts := options.Find().SetSort(bson.D{{Key: "waktu", Value: -1}})
	cursor, err := r.kehadiran.Find(ctx, bson.M{"sesi_id": sesiID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	list := []model.KehadiranManasik{}
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// MarkHadir records attendance. If the jamaah is already recorded for this
// session, it returns the existing record and created=false.
func (r *ManasikRepository) MarkHadir(ctx context.Context, k *model.KehadiranManasik) (existing *model.KehadiranManasik, created bool, err error) {
	k.Waktu = time.Now()
	res, err := r.kehadiran.InsertOne(ctx, k)
	if err == nil {
		k.ID = res.InsertedID.(primitive.ObjectID)
		return k, true, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return nil, false, err
	}
	var prev model.KehadiranManasik
	if err := r.kehadiran.FindOne(ctx, bson.M{"sesi_id": k.SesiID, "jamaah_id": k.JamaahID}).Decode(&prev); err != nil {
		return nil, false, err
	}
	return &prev, false, nil
}

func (r *ManasikRepository) DeleteKehadiran(ctx context.Context, sesiID, jamaahID primitive.ObjectID) (bool, error) {
	res, err := r.kehadiran.DeleteOne(ctx, bson.M{"sesi_id": sesiID, "jamaah_id": jamaahID})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}
