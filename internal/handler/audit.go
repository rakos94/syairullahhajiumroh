package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func logAudit(c *gin.Context, auditRepo *repository.AuditLogRepository, entityType string, entityID primitive.ObjectID, entityLabel, action, description string) {
	logAuditWithChanges(c, auditRepo, entityType, entityID, entityLabel, action, description, nil)
}

func logAuditWithChanges(c *gin.Context, auditRepo *repository.AuditLogRepository, entityType string, entityID primitive.ObjectID, entityLabel, action, description string, changes []model.FieldChange) {
	adminIDStr, _ := c.Get("admin_id")
	adminUsername, _ := c.Get("admin_username")

	var adminID primitive.ObjectID
	if idStr, ok := adminIDStr.(string); ok {
		adminID, _ = primitive.ObjectIDFromHex(idStr)
	}
	var username string
	if u, ok := adminUsername.(string); ok {
		username = u
	}

	entry := &model.AuditLog{
		AdminID:       adminID,
		AdminUsername: username,
		EntityType:    entityType,
		EntityID:      entityID,
		EntityLabel:   entityLabel,
		Action:        action,
		Description:   description,
		Changes:       changes,
	}

	if err := auditRepo.Create(context.Background(), entry); err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}

// diffField appends a change if old != new.
func diffField(changes *[]model.FieldChange, field, oldVal, newVal string) {
	if oldVal != newVal {
		*changes = append(*changes, model.FieldChange{Field: field, OldValue: oldVal, NewValue: newVal})
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%02d/%02d/%04d", t.Day(), t.Month(), t.Year())
}

func formatBool(b bool) string {
	if b {
		return "Ya"
	}
	return "Tidak"
}

var statusLabels = map[string]string{
	"belum_bayar": "Belum Bayar",
	"dp":          "DP",
	"lunas":       "Lunas",
}

func diffJamaah(old, new *model.Jamaah) []model.FieldChange {
	var changes []model.FieldChange
	diffField(&changes, "Nama", old.Nama, new.Nama)
	diffField(&changes, "NIK", old.NIK, new.NIK)
	diffField(&changes, "Nomor Paspor", old.NomorPaspor, new.NomorPaspor)
	diffField(&changes, "Alamat", old.Alamat, new.Alamat)
	diffField(&changes, "No HP", old.NoHP, new.NoHP)
	diffField(&changes, "Tempat Lahir", old.TempatLahir, new.TempatLahir)
	diffField(&changes, "Tanggal Lahir", formatDate(old.TanggalLahir), formatDate(new.TanggalLahir))
	diffField(&changes, "Jenis Kelamin", old.JenisKelamin, new.JenisKelamin)

	if old.PaketID != new.PaketID {
		changes = append(changes, model.FieldChange{Field: "Paket", OldValue: old.PaketID.Hex(), NewValue: new.PaketID.Hex()})
	}

	oldTK := ""
	if old.TanggalKeberangkatan != nil {
		oldTK = old.TanggalKeberangkatan.Nama
	}
	newTK := ""
	if new.TanggalKeberangkatan != nil {
		newTK = new.TanggalKeberangkatan.Nama
	}
	diffField(&changes, "Tanggal Keberangkatan", oldTK, newTK)

	oldStatus := statusLabels[old.StatusPembayaran]
	if oldStatus == "" {
		oldStatus = old.StatusPembayaran
	}
	newStatus := statusLabels[new.StatusPembayaran]
	if newStatus == "" {
		newStatus = new.StatusPembayaran
	}
	diffField(&changes, "Status Pembayaran", oldStatus, newStatus)

	diffField(&changes, "No Rekening Haji", old.NoRekeningHaji, new.NoRekeningHaji)
	diffField(&changes, "Tipe Bank", old.TipeBank, new.TipeBank)
	diffField(&changes, "Batik Nasional Sudah Dijahit", formatBool(old.BatikNasionalSudahDijahit), formatBool(new.BatikNasionalSudahDijahit))
	diffField(&changes, "Batik KBIH Sudah Diterima", formatBool(old.BatikKBIHSudahDiterima), formatBool(new.BatikKBIHSudahDiterima))
	diffField(&changes, "Koper Sudah Diterima", formatBool(old.KoperSudahDiterima), formatBool(new.KoperSudahDiterima))
	diffField(&changes, "Keterangan", old.Keterangan, new.Keterangan)

	return changes
}

func diffPaket(old, new *model.Paket) []model.FieldChange {
	var changes []model.FieldChange
	diffField(&changes, "Tipe", old.Tipe, new.Tipe)
	diffField(&changes, "Tahun", fmt.Sprintf("%d", old.Tahun), fmt.Sprintf("%d", new.Tahun))
	diffField(&changes, "Bulan", fmt.Sprintf("%d", old.Bulan), fmt.Sprintf("%d", new.Bulan))
	return changes
}
