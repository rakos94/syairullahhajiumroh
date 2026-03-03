package handler

import (
	"net/http"

	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	jamaahRepo *repository.JamaahRepository
	paketRepo  *repository.PaketRepository
}

func NewStatisticsHandler(jamaahRepo *repository.JamaahRepository, paketRepo *repository.PaketRepository) *StatisticsHandler {
	return &StatisticsHandler{jamaahRepo: jamaahRepo, paketRepo: paketRepo}
}

func (h *StatisticsHandler) GetStatistics(c *gin.Context) {
	stats, err := h.jamaahRepo.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Resolve paket labels
	type PaketStat struct {
		PaketID    string `json:"paket_id"`
		Label      string `json:"label"`
		Tipe       string `json:"tipe"`
		Total      int64  `json:"total"`
		BelumBayar int64  `json:"belum_bayar"`
		DP         int64  `json:"dp"`
		Lunas      int64  `json:"lunas"`
	}

	var paketStats []PaketStat
	var totalHaji, totalUmroh int64

	if len(stats.PaketBreakdown) > 0 {
		// Fetch all pakets to build labels
		allPakets, _ := h.paketRepo.FindAll(c.Request.Context(), "")
		paketMap := make(map[string]struct {
			Label string
			Tipe  string
		})
		for _, p := range allPakets {
			p.BuildLabel()
			paketMap[p.ID.Hex()] = struct {
				Label string
				Tipe  string
			}{p.Label, p.Tipe}
		}

		for _, p := range stats.PaketBreakdown {
			info := paketMap[p.PaketID.Hex()]
			paketStats = append(paketStats, PaketStat{
				PaketID:    p.PaketID.Hex(),
				Label:      info.Label,
				Tipe:       info.Tipe,
				Total:      p.Total,
				BelumBayar: p.BelumBayar,
				DP:         p.DP,
				Lunas:      p.Lunas,
			})
			if info.Tipe == "haji" {
				totalHaji += p.Total
			} else {
				totalUmroh += p.Total
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":               stats.Total,
		"total_haji":          totalHaji,
		"total_umroh":         totalUmroh,
		"payment_breakdown":   stats.PaymentBreakdown,
		"paket_breakdown":     paketStats,
		"total_haji_completion": stats.TotalHaji,
		"batik_nasional_done":  stats.BatikNasionalDone,
		"batik_kbih_done":      stats.BatikKBIHDone,
		"koper_done":           stats.KoperDone,
	})
}
