package handler

import (
	"fmt"
	"net/http"
	"strings"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxBulkManasik = 500

type bulkManasikRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
	// nil = leave unchanged, "" = clear
	GedungManasik *string `json:"gedung_manasik"`
	KamarManasik  *string `json:"kamar_manasik"`
	Rombongan     *string `json:"rombongan"`
	Regu          *string `json:"regu"`
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	return &v
}

// manasikChanges returns the audit diff of applying u to old.
func manasikChanges(old *model.Jamaah, u repository.ManasikUpdate) []model.FieldChange {
	var changes []model.FieldChange
	if u.GedungManasik != nil {
		diffField(&changes, "Gedung Manasik", old.GedungManasik, *u.GedungManasik)
	}
	if u.KamarManasik != nil {
		diffField(&changes, "Kamar Manasik", old.KamarManasik, *u.KamarManasik)
	}
	if u.Rombongan != nil {
		diffField(&changes, "Rombongan", old.Rombongan, *u.Rombongan)
	}
	if u.Regu != nil {
		diffField(&changes, "Regu", old.Regu, *u.Regu)
	}
	return changes
}

// BulkManasik godoc
// @Summary      Atur gedung/kamar/rombongan/regu manasik untuk banyak jamaah
// @Description  Hanya jamaah dengan paket haji. Field yang tidak dikirim tidak diubah; string kosong menghapus nilai.
// @Tags         jamaah
// @Accept       json
// @Produce      json
// @Param        body  body      bulkManasikRequest  true  "ID jamaah dan penempatan"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /jamaah/bulk-manasik [put]
func (h *JamaahHandler) BulkManasik(c *gin.Context) {
	var req bulkManasikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pilih minimal satu jamaah"})
		return
	}
	u := repository.ManasikUpdate{
		GedungManasik: trimPtr(req.GedungManasik),
		KamarManasik:  trimPtr(req.KamarManasik),
		Rombongan:     trimPtr(req.Rombongan),
		Regu:          trimPtr(req.Regu),
	}
	if u.IsEmpty() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isi minimal satu: gedung, kamar, rombongan, atau regu"})
		return
	}
	if len(req.IDs) > maxBulkManasik {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("maksimal %d jamaah sekaligus", maxBulkManasik)})
		return
	}

	ids := make([]primitive.ObjectID, 0, len(req.IDs))
	for _, s := range req.IDs {
		oid, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid: " + s})
			return
		}
		ids = append(ids, oid)
	}

	ctx := c.Request.Context()
	existing, err := h.repo.FindByIDs(ctx, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
		return
	}
	h.populatePaket(c, existing)

	// Manasik only applies to haji jamaah; refuse the whole batch otherwise
	// so the admin notices a wrong selection instead of a partial update.
	var nonHaji []string
	for _, j := range existing {
		if j.Paket == nil || j.Paket.Tipe != "haji" {
			nonHaji = append(nonHaji, j.Nama)
		}
	}
	if len(nonHaji) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manasik hanya untuk jamaah haji: " + strings.Join(nonHaji, ", ")})
		return
	}

	foundIDs := make([]primitive.ObjectID, len(existing))
	for i, j := range existing {
		foundIDs[i] = j.ID
	}
	if _, err := h.repo.SetManasik(ctx, foundIDs, u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range existing {
		old := &existing[i]
		if changes := manasikChanges(old, u); len(changes) > 0 {
			logAuditWithChanges(c, h.auditRepo, "jamaah", old.ID, old.Nama, "update",
				fmt.Sprintf("Mengatur penempatan manasik (massal): %s", old.Nama), changes)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d jamaah berhasil diperbarui", len(existing)),
		"updated": len(existing),
	})
}
