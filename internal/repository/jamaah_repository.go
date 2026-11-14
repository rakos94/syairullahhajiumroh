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

func (r *JamaahRepository) FindAll(ctx context.Context, paketID *primitive.ObjectID, search, statusPembayaran, kelengkapan string, page, limit int) ([]model.Jamaah, int64, error) {
	filter := bson.M{"deleted_at": nil}
	if paketID != nil {
		filter["paket_id"] = *paketID
	}
	if statusPembayaran != "" {
		filter["status_pembayaran"] = statusPembayaran
	}
	switch kelengkapan {
	case "batik_nasional_belum":
		filter["batik_nasional_sudah_dijahit"] = bson.M{"$ne": true}
	case "batik_kbih_belum":
		filter["batik_kbih_sudah_diterima"] = bson.M{"$ne": true}
	case "koper_belum":
		filter["koper_sudah_diterima"] = bson.M{"$ne": true}
	case "manasik_belum":
		// Only haji jamaah do manasik; list those still missing gedung or kamar.
		hajiIDs, err := r.collection.Database().Collection("paket").Distinct(ctx, "_id", bson.M{"tipe": "haji"})
		if err != nil {
			return nil, 0, err
		}
		filter["$and"] = bson.A{
			bson.M{"paket_id": bson.M{"$in": hajiIDs}},
			bson.M{"$or": bson.A{
				bson.M{"gedung_manasik": bson.M{"$in": bson.A{"", nil}}},
				bson.M{"kamar_manasik": bson.M{"$in": bson.A{"", nil}}},
			}},
		}
	}
	if search != "" {
		regex := bson.M{"$regex": search, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"nama": regex},
			bson.M{"nik": bson.M{"$regex": search}},
		}
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

func (r *JamaahRepository) FindByNIK(ctx context.Context, nik string) (*model.Jamaah, error) {
	var jamaah model.Jamaah
	err := r.collection.FindOne(ctx, bson.M{"nik": nik, "deleted_at": nil}).Decode(&jamaah)
	if err != nil {
		return nil, err
	}
	return &jamaah, nil
}

func (r *JamaahRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.Jamaah, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}, "deleted_at": nil})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Jamaah
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ManasikUpdate holds manasik placement fields. nil = leave unchanged, "" = clear.
type ManasikUpdate struct {
	GedungManasik *string
	KamarManasik  *string
	Rombongan     *string
	Regu          *string
}

func (u ManasikUpdate) IsEmpty() bool {
	return u.GedungManasik == nil && u.KamarManasik == nil && u.Rombongan == nil && u.Regu == nil
}

func (u ManasikUpdate) setDoc() bson.M {
	set := bson.M{"updated_at": time.Now()}
	if u.GedungManasik != nil {
		set["gedung_manasik"] = *u.GedungManasik
	}
	if u.KamarManasik != nil {
		set["kamar_manasik"] = *u.KamarManasik
	}
	if u.Rombongan != nil {
		set["rombongan"] = *u.Rombongan
	}
	if u.Regu != nil {
		set["regu"] = *u.Regu
	}
	return set
}

// SetManasik applies the same manasik placement to many jamaah at once.
func (r *JamaahRepository) SetManasik(ctx context.Context, ids []primitive.ObjectID, u ManasikUpdate) (int64, error) {
	res, err := r.collection.UpdateMany(
		ctx,
		bson.M{"_id": bson.M{"$in": ids}, "deleted_at": nil},
		bson.M{"$set": u.setDoc()},
	)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

// SetManasikOne applies a manasik placement to a single jamaah (used by import).
func (r *JamaahRepository) SetManasikOne(ctx context.Context, id primitive.ObjectID, u ManasikUpdate) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "deleted_at": nil}, bson.M{"$set": u.setDoc()})
	return err
}

// FindByPaket returns all non-deleted jamaah of a paket, sorted by name.
func (r *JamaahRepository) FindByPaket(ctx context.Context, paketID primitive.ObjectID) ([]model.Jamaah, error) {
	opts := options.Find().SetSort(bson.D{{Key: "nama", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"paket_id": paketID, "deleted_at": nil}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	results := []model.Jamaah{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindActive returns all non-deleted jamaah, optionally of one paket, sorted by name.
func (r *JamaahRepository) FindActive(ctx context.Context, paketID *primitive.ObjectID) ([]model.Jamaah, error) {
	filter := bson.M{"deleted_at": nil}
	if paketID != nil {
		filter["paket_id"] = *paketID
	}
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "nama", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	results := []model.Jamaah{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// FindLeaders returns the ketua regu of (paket, rombongan, regu) and the
// ketua rombongan of (paket, rombongan). Empty rombongan/regu yields nothing.
func (r *JamaahRepository) FindLeaders(ctx context.Context, paketID primitive.ObjectID, rombongan, regu string) (ketuaRegu, ketuaRombongan []model.Jamaah, err error) {
	find := func(filter bson.M) ([]model.Jamaah, error) {
		filter["paket_id"] = paketID
		filter["deleted_at"] = nil
		cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "nama", Value: 1}}))
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		out := []model.Jamaah{}
		err = cursor.All(ctx, &out)
		return out, err
	}
	if rombongan != "" && regu != "" {
		if ketuaRegu, err = find(bson.M{"rombongan": rombongan, "regu": regu, "ketua_regu": true}); err != nil {
			return nil, nil, err
		}
	}
	if rombongan != "" {
		if ketuaRombongan, err = find(bson.M{"rombongan": rombongan, "ketua_rombongan": true}); err != nil {
			return nil, nil, err
		}
	}
	return ketuaRegu, ketuaRombongan, nil
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

func (r *JamaahRepository) PushToArray(ctx context.Context, id primitive.ObjectID, field string, value interface{}) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deleted_at": nil},
		bson.M{
			"$push": bson.M{field: value},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *JamaahRepository) UpdateDepartureByPaket(ctx context.Context, paketID primitive.ObjectID, nama string, tanggal *time.Time) error {
	filter := bson.M{
		"paket_id":                   paketID,
		"deleted_at":                 nil,
		"tanggal_keberangkatan.nama": nama,
	}
	update := bson.M{"$set": bson.M{
		"tanggal_keberangkatan.tanggal": tanggal,
		"updated_at":                    time.Now(),
	}}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *JamaahRepository) PullFromArray(ctx context.Context, id primitive.ObjectID, field string, value interface{}) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deleted_at": nil},
		bson.M{
			"$pull": bson.M{field: value},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

type PaymentCount struct {
	Status string `bson:"_id" json:"status"`
	Count  int64  `bson:"count" json:"count"`
}

type PaketCount struct {
	PaketID    primitive.ObjectID `bson:"_id" json:"paket_id"`
	Total      int64              `bson:"total" json:"total"`
	BelumBayar int64              `bson:"belum_bayar" json:"belum_bayar"`
	DP         int64              `bson:"dp" json:"dp"`
	Lunas      int64              `bson:"lunas" json:"lunas"`
}

type Statistics struct {
	Total              int64          `json:"total"`
	PaymentBreakdown   []PaymentCount `json:"payment_breakdown"`
	PaketBreakdown     []PaketCount   `json:"paket_breakdown"`
	TotalHaji          int64          `json:"total_haji"`
	BatikNasionalDone  int64          `json:"batik_nasional_done"`
	BatikKBIHDone      int64          `json:"batik_kbih_done"`
	KoperDone          int64          `json:"koper_done"`
}

func (r *JamaahRepository) GetStatistics(ctx context.Context) (*Statistics, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"deleted_at": nil}}},
		bson.D{{Key: "$facet", Value: bson.M{
			"total": bson.A{
				bson.M{"$count": "count"},
			},
			"by_payment": bson.A{
				bson.M{"$group": bson.M{
					"_id":   "$status_pembayaran",
					"count": bson.M{"$sum": 1},
				}},
			},
			"by_paket": bson.A{
				bson.M{"$group": bson.M{
					"_id":   "$paket_id",
					"total": bson.M{"$sum": 1},
					"belum_bayar": bson.M{"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$status_pembayaran", "belum_bayar"}}, 1, 0},
					}},
					"dp": bson.M{"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$status_pembayaran", "dp"}}, 1, 0},
					}},
					"lunas": bson.M{"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$status_pembayaran", "lunas"}}, 1, 0},
					}},
				}},
			},
			"completion": bson.A{
				bson.M{"$lookup": bson.M{
					"from":         "paket",
					"localField":   "paket_id",
					"foreignField": "_id",
					"as":           "paket_info",
				}},
				bson.M{"$match": bson.M{
					"paket_info.tipe": "haji",
				}},
				bson.M{"$group": bson.M{
					"_id":        nil,
					"total_haji": bson.M{"$sum": 1},
					"batik_nasional": bson.M{"$sum": bson.M{
						"$cond": bson.A{"$batik_nasional_sudah_dijahit", 1, 0},
					}},
					"batik_kbih": bson.M{"$sum": bson.M{
						"$cond": bson.A{"$batik_kbih_sudah_diterima", 1, 0},
					}},
					"koper": bson.M{"$sum": bson.M{
						"$cond": bson.A{"$koper_sudah_diterima", 1, 0},
					}},
				}},
			},
		}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := &Statistics{}
	if len(results) == 0 {
		return stats, nil
	}

	facet := results[0]

	// Total
	if totals, ok := facet["total"].(bson.A); ok && len(totals) > 0 {
		if doc, ok := totals[0].(bson.M); ok {
			if v, ok := doc["count"].(int32); ok {
				stats.Total = int64(v)
			} else if v, ok := doc["count"].(int64); ok {
				stats.Total = v
			}
		}
	}

	// Payment breakdown
	if payments, ok := facet["by_payment"].(bson.A); ok {
		for _, p := range payments {
			if doc, ok := p.(bson.M); ok {
				pc := PaymentCount{}
				if s, ok := doc["_id"].(string); ok {
					pc.Status = s
				}
				if v, ok := doc["count"].(int32); ok {
					pc.Count = int64(v)
				} else if v, ok := doc["count"].(int64); ok {
					pc.Count = v
				}
				stats.PaymentBreakdown = append(stats.PaymentBreakdown, pc)
			}
		}
	}

	// Paket breakdown
	if pakets, ok := facet["by_paket"].(bson.A); ok {
		for _, p := range pakets {
			if doc, ok := p.(bson.M); ok {
				pk := PaketCount{}
				if id, ok := doc["_id"].(primitive.ObjectID); ok {
					pk.PaketID = id
				}
				pk.Total = toInt64(doc["total"])
				pk.BelumBayar = toInt64(doc["belum_bayar"])
				pk.DP = toInt64(doc["dp"])
				pk.Lunas = toInt64(doc["lunas"])
				stats.PaketBreakdown = append(stats.PaketBreakdown, pk)
			}
		}
	}

	// Completion (haji only)
	if comp, ok := facet["completion"].(bson.A); ok && len(comp) > 0 {
		if doc, ok := comp[0].(bson.M); ok {
			stats.TotalHaji = toInt64(doc["total_haji"])
			stats.BatikNasionalDone = toInt64(doc["batik_nasional"])
			stats.BatikKBIHDone = toInt64(doc["batik_kbih"])
			stats.KoperDone = toInt64(doc["koper"])
		}
	}

	return stats, nil
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int32:
		return int64(n)
	case int64:
		return n
	default:
		return 0
	}
}
