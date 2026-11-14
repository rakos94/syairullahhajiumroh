package migration

import (
	"context"
	"fmt"
	"log"
	"time"

	"syairullahhajiumroh/internal/model"

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

func statusLabelMigrasi(s string) string {
	switch s {
	case "belum_bayar":
		return "Belum Bayar"
	case "dp":
		return "DP"
	case "lunas":
		return "Lunas"
	}
	return s
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
		{
			Name: "006_create_audit_logs_indexes",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("audit_logs")
				_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
					{
						Keys: bson.D{
							{Key: "entity_type", Value: 1},
							{Key: "entity_id", Value: 1},
							{Key: "created_at", Value: -1},
						},
					},
					{
						Keys: bson.D{{Key: "created_at", Value: -1}},
					},
					{
						Keys: bson.D{{Key: "admin_id", Value: 1}},
					},
				})
				return err
			},
		},
		{
			Name: "007_create_pengumuman_indexes",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				coll := db.Collection("pengumuman")
				_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{
						{Key: "dipublikasikan", Value: 1},
						{Key: "disematkan", Value: -1},
						{Key: "created_at", Value: -1},
					},
				})
				return err
			},
		},
		{
			Name: "008_create_manasik_indexes",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				if _, err := db.Collection("sesi_manasik").Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{{Key: "paket_id", Value: 1}, {Key: "tanggal", Value: 1}},
				}); err != nil {
					return err
				}
				_, err := db.Collection("kehadiran_manasik").Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys:    bson.D{{Key: "sesi_id", Value: 1}, {Key: "jamaah_id", Value: 1}},
					Options: options.Index().SetUnique(true),
				})
				return err
			},
		},
		{
			Name: "009_status_pembayaran_otomatis",
			ApplyFn: func(ctx context.Context, db *mongo.Database) error {
				// Payment status is now derived from receipts for priced pakets.
				pakets := map[primitive.ObjectID]*model.Paket{}
				cur, err := db.Collection("paket").Find(ctx, bson.M{})
				if err != nil {
					return err
				}
				var list []model.Paket
				if err := cur.All(ctx, &list); err != nil {
					return err
				}
				for i := range list {
					pakets[list[i].ID] = &list[i]
				}
				jc := db.Collection("jamaah")
				cur, err = jc.Find(ctx, bson.M{"deleted_at": nil})
				if err != nil {
					return err
				}
				var jamaah []model.Jamaah
				if err := cur.All(ctx, &jamaah); err != nil {
					return err
				}
				changed := 0
				for i := range jamaah {
					j := &jamaah[i]
					status, ok := model.StatusOtomatis(model.HitungTagihan(j, pakets[j.PaketID]))
					if !ok || status == j.StatusPembayaran {
						continue
					}
					if _, err := jc.UpdateOne(ctx, bson.M{"_id": j.ID}, bson.M{"$set": bson.M{"status_pembayaran": status}}); err != nil {
						return err
					}
					// Record each change in Riwayat so admins can review it
					// (e.g. "Lunas" set by hand without an uploaded receipt).
					if _, err := db.Collection("audit_logs").InsertOne(ctx, model.AuditLog{
						AdminUsername: "sistem",
						EntityType:    "jamaah",
						EntityID:      j.ID,
						EntityLabel:   j.Nama,
						Action:        "update",
						Description:   fmt.Sprintf("Status pembayaran disesuaikan dengan bukti pembayaran (otomatis): %s", j.Nama),
						Changes: []model.FieldChange{{
							Field: "Status Pembayaran", OldValue: statusLabelMigrasi(j.StatusPembayaran), NewValue: statusLabelMigrasi(status),
						}},
						CreatedAt: time.Now(),
					}); err != nil {
						return err
					}
					log.Printf("  status pembayaran %s: %s -> %s", j.Nama, j.StatusPembayaran, status)
					changed++
				}
				log.Printf("  %d jamaah diperbarui", changed)
				return nil
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
