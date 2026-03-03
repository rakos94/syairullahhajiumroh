package migration

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MigrationRecord struct {
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}

type Migration struct {
	Name    string
	ApplyFn func(ctx context.Context, db *mongo.Database) error
}

func RunMigrations(ctx context.Context, db *mongo.Database) error {
	migrationsColl := db.Collection("migrations")

	migrations := []Migration{
		{
			Name: "001_create_jamaah_indexes",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("jamaah")
				indexes := []mongo.IndexModel{
					{
						Keys:    bson.D{{Key: "nik", Value: 1}},
						Options: options.Index().SetUnique(true),
					},
					{
						Keys:    bson.D{{Key: "nomor_paspor", Value: 1}},
						Options: options.Index().SetUnique(true).SetSparse(true),
					},
					{
						Keys: bson.D{{Key: "paket", Value: 1}},
					},
					{
						Keys: bson.D{{Key: "status_pembayaran", Value: 1}},
					},
				}
				_, err := coll.Indexes().CreateMany(ctx, indexes)
				return err
			},
		},
		{
			Name: "002_add_deleted_at_index",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("jamaah")
				_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{{Key: "deleted_at", Value: 1}},
				})
				return err
			},
		},
		{
			Name: "003_create_paket_indexes",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("paket")
				_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{
						{Key: "tipe", Value: 1},
						{Key: "tahun", Value: 1},
						{Key: "bulan", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				})
				if err != nil {
					return err
				}

				// Add paket_id index on jamaah
				jamaahColl := db.Collection("jamaah")
				_, err = jamaahColl.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{{Key: "paket_id", Value: 1}},
				})
				return err
			},
		},
		{
			Name: "004_migrate_tanggal_keberangkatan",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("jamaah")
				cursor, err := coll.Find(ctx, bson.M{
					"tanggal_keberangkatan": bson.M{"$type": "date"},
				})
				if err != nil {
					return err
				}
				defer cursor.Close(ctx)

				for cursor.Next(ctx) {
					var doc bson.M
					if err := cursor.Decode(&doc); err != nil {
						return err
					}
					id := doc["_id"]
					oldDate := doc["tanggal_keberangkatan"].(primitive.DateTime)
					_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
						"$set": bson.M{"tanggal_keberangkatan": bson.M{"tanggal": oldDate}},
					})
					if err != nil {
						return err
					}
				}
				return cursor.Err()
			},
		},
		{
			Name: "005_create_admins_index",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("admins")
				_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys:    bson.D{{Key: "username", Value: 1}},
					Options: options.Index().SetUnique(true),
				})
				return err
			},
		},
	}

	for _, m := range migrations {
		var record MigrationRecord
		err := migrationsColl.FindOne(ctx, bson.M{"name": m.Name}).Decode(&record)
		if err == nil {
			log.Printf("Migration %s already applied, skipping", m.Name)
			continue
		}

		log.Printf("Applying migration: %s", m.Name)
		if err := m.ApplyFn(ctx, db); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Name, err)
		}

		_, err = migrationsColl.InsertOne(ctx, MigrationRecord{
			Name:      m.Name,
			AppliedAt: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.Name, err)
		}

		log.Printf("Migration %s applied successfully", m.Name)
	}

	return nil
}
