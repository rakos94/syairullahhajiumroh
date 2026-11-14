package handler

import (
	"net/http"
	"strings"
	"time"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// PortalHandler serves the public, self-service lookup used by jamaah
// to check their own manasik room assignment.
type PortalHandler struct {
	repo      *repository.JamaahRepository
	paketRepo *repository.PaketRepository
}

func NewPortalHandler(repo *repository.JamaahRepository, paketRepo *repository.PaketRepository) *PortalHandler {
	return &PortalHandler{repo: repo, paketRepo: paketRepo}
}

type cekKamarRequest struct {
	NIK          string `json:"nik" binding:"required"`
	TanggalLahir string `json:"tanggal_lahir" binding:"required"` // yyyy-mm-dd
}

// CekKamarResponse only exposes what a jamaah needs to see; never return the
// full Jamaah document from this public endpoint.
type CekKamarResponse struct {
	Nama                 string                     `json:"nama"`
	Paket                string                     `json:"paket"`
	TanggalKeberangkatan *model.KeberangkatanJamaah `json:"tanggal_keberangkatan,omitempty"`
	GedungManasik        string                     `json:"gedung_manasik"`
	KamarManasik         string                     `json:"kamar_manasik"`
}

// CekKamar godoc
// @Summary      Cek kamar manasik (publik)
// @Description  Jamaah memasukkan NIK dan tanggal lahir untuk melihat gedung dan kamar manasik
// @Tags         portal
// @Accept       json
// @Produce      json
// @Param        body  body      cekKamarRequest  true  "NIK dan tanggal lahir (yyyy-mm-dd)"
// @Success      200   {object}  CekKamarResponse
// @Failure      404   {object}  map[string]string
// @Router       /portal/cek-kamar [post]
func (h *PortalHandler) CekKamar(c *gin.Context) {
	var req cekKamarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NIK dan tanggal lahir wajib diisi"})
		return
	}

	lahir, err := time.Parse("2006-01-02", strings.TrimSpace(req.TanggalLahir))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format tanggal lahir tidak valid"})
		return
	}

	// Same message for unknown NIK and wrong birth date, so the endpoint
	// can't be used to check whether a NIK is registered.
	notFound := gin.H{"error": "data tidak ditemukan, periksa kembali NIK dan tanggal lahir"}

	jamaah, err := h.repo.FindByNIK(c.Request.Context(), strings.TrimSpace(req.NIK))
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan, coba lagi nanti"})
		return
	}

	if jamaah.TanggalLahir.IsZero() || jamaah.TanggalLahir.UTC().Format("2006-01-02") != lahir.Format("2006-01-02") {
		c.JSON(http.StatusNotFound, notFound)
		return
	}

	resp := CekKamarResponse{
		Nama:                 jamaah.Nama,
		TanggalKeberangkatan: jamaah.TanggalKeberangkatan,
		GedungManasik:        jamaah.GedungManasik,
		KamarManasik:         jamaah.KamarManasik,
	}
	if !jamaah.PaketID.IsZero() {
		if paket, err := h.paketRepo.FindByID(c.Request.Context(), jamaah.PaketID); err == nil {
			resp.Paket = paket.Label
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PortalHandler) RegisterRoutes(r *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	handlers := append(middlewares, h.CekKamar)
	r.POST("/portal/cek-kamar", handlers...)
}
