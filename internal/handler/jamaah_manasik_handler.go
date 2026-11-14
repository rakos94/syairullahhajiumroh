package handler

import (
	"fmt"
	"net/http"
	"strings"

	"syairullahhajiumroh/internal/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxBulkManasik = 500

type bulkManasikRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
	// nil = leave unchanged, "" = clear
	GedungManasik *string `json:"gedung_manasik"`
	KamarManasik  *string `json:"kamar_manasik"`
}

// BulkManasik godoc
// @Summary      Atur gedung/kamar manasik untuk banyak jamaah
// @Description  Hanya jamaah dengan paket haji yang diubah. Field yang tidak dikirim tidak diubah; string kosong menghapus nilai.
// @Tags         jamaah
// @Accept       json
// @Produce      json
// @Param        body  body      bulkManasikRequest  true  "ID jamaah dan gedung/kamar"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /jamaah/bulk-manasik [put]
func (h *JamaahHandler) BulkManasik(c *gin.Context) {
	var req bulkManasikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pilih minimal satu jamaah"})
		return
	}
	if req.GedungManasik == nil && req.KamarManasik == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isi gedung atau kamar manasik"})
		return
	}
	if len(req.IDs) > maxBulkManasik {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("maksimal %d jamaah sekaligus", maxBulkManasik)})
		return
	}
	if req.GedungManasik != nil {
		v := strings.TrimSpace(*req.GedungManasik)
		req.GedungManasik = &v
	}
	if req.KamarManasik != nil {
		v := strings.TrimSpace(*req.KamarManasik)
		req.KamarManasik = &v
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
	if len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
		return
	}

	foundIDs := make([]primitive.ObjectID, len(existing))
	for i, j := range existing {
		foundIDs[i] = j.ID
	}
	if _, err := h.repo.SetManasik(ctx, foundIDs, req.GedungManasik, req.KamarManasik); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, old := range existing {
		var changes []model.FieldChange
		if req.GedungManasik != nil {
			diffField(&changes, "Gedung Manasik", old.GedungManasik, *req.GedungManasik)
		}
		if req.KamarManasik != nil {
			diffField(&changes, "Kamar Manasik", old.KamarManasik, *req.KamarManasik)
		}
		if len(changes) > 0 {
			logAuditWithChanges(c, h.auditRepo, "jamaah", old.ID, old.Nama, "update",
				fmt.Sprintf("Mengatur kamar manasik (massal): %s", old.Nama), changes)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d jamaah berhasil diperbarui", len(existing)),
		"updated": len(existing),
	})
}
