package handler

import (
	"fmt"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
)

// applyStatusOtomatis sets j.StatusPembayaran from its receipts when the
// paket has a price. Without a price the status stays as entered.
func applyStatusOtomatis(j *model.Jamaah, p *model.Paket) {
	if status, ok := model.StatusOtomatis(model.HitungTagihan(j, p)); ok {
		j.StatusPembayaran = status
	}
}

// syncStatusPembayaran recomputes and stores the status of an existing jamaah
// (after receipts or prices change) and audits the change.
func syncStatusPembayaran(c *gin.Context, jamaahRepo *repository.JamaahRepository, auditRepo *repository.AuditLogRepository, j *model.Jamaah, p *model.Paket) {
	status, ok := model.StatusOtomatis(model.HitungTagihan(j, p))
	if !ok || status == j.StatusPembayaran {
		return
	}
	if err := jamaahRepo.UpdateField(c.Request.Context(), j.ID, "status_pembayaran", status); err != nil {
		return
	}
	var changes []model.FieldChange
	diffField(&changes, "Status Pembayaran", statusLabelOf(j.StatusPembayaran), statusLabelOf(status))
	logAuditWithChanges(c, auditRepo, "jamaah", j.ID, j.Nama, "update",
		fmt.Sprintf("Status pembayaran otomatis diperbarui: %s", j.Nama), changes)
	j.StatusPembayaran = status
}

func statusLabelOf(s string) string {
	if l := statusLabels[s]; l != "" {
		return l
	}
	return s
}
