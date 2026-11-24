package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
)

// PortalHandler serves the public, self-service lookup used by jamaah
// to check their own manasik room assignment.
type PortalHandler struct {
	repo      *repository.JamaahRepository
	paketRepo *repository.PaketRepository
	qrSecret  string
}

func NewPortalHandler(repo *repository.JamaahRepository, paketRepo *repository.PaketRepository, qrSecret string) *PortalHandler {
	return &PortalHandler{repo: repo, paketRepo: paketRepo, qrSecret: qrSecret}
}

// StatusItem is a yes/no checklist entry shown to the jamaah.
type StatusItem struct {
	Nama    string `json:"nama"`
	Selesai bool   `json:"selesai"`
}

// KontakKetua is the public contact of a group leader.
type KontakKetua struct {
	Nama string `json:"nama"`
	NoHP string `json:"no_hp"`
}

func toKontak(list []model.Jamaah) []KontakKetua {
	out := make([]KontakKetua, 0, len(list))
	for _, j := range list {
		out = append(out, KontakKetua{Nama: j.Nama, NoHP: j.NoHP})
	}
	return out
}

type cekKamarRequest struct {
	NIK          string `json:"nik" binding:"required"`
	TanggalLahir string `json:"tanggal_lahir" binding:"required"` // yyyy-mm-dd
}

// CekKamarResponse only exposes what a jamaah needs to see; never return the
// full Jamaah document from this public endpoint. One person (NIK) can be
// registered in several pakets, so each registration is one entry.
type CekKamarResponse struct {
	Nama  string          `json:"nama"`
	Paket []CekKamarPaket `json:"paket"`
}

// CekKamarPaket is one registration (paket) of the jamaah.
type CekKamarPaket struct {
	Paket                string                     `json:"paket"`
	TipePaket            string                     `json:"tipe_paket"`
	TanggalKeberangkatan *model.KeberangkatanJamaah `json:"tanggal_keberangkatan,omitempty"`
	GedungManasik        string                     `json:"gedung_manasik"`
	KamarManasik         string                     `json:"kamar_manasik"`
	Rombongan            string                     `json:"rombongan"`
	Regu                 string                     `json:"regu"`
	KetuaRegu            []KontakKetua              `json:"ketua_regu"`
	KetuaRombongan       []KontakKetua              `json:"ketua_rombongan"`
	// Hotel names in Mekkah/Madinah (per paket) and the jamaah's room numbers.
	HotelMekkah       string `json:"hotel_mekkah,omitempty"`
	HotelMadinah      string `json:"hotel_madinah,omitempty"`
	KamarHotelMekkah  string `json:"kamar_hotel_mekkah,omitempty"`
	KamarHotelMadinah string `json:"kamar_hotel_madinah,omitempty"`
	MapsHotelMekkah   string `json:"maps_hotel_mekkah,omitempty"`
	MapsHotelMadinah  string `json:"maps_hotel_madinah,omitempty"`
	// Kelengkapan and Dokumen only say whether each item is done/uploaded;
	// the documents themselves are never exposed here. Payment status is
	// deliberately left out (NIK + birth date is a weak credential).
	Kelengkapan []StatusItem `json:"kelengkapan"`
	Dokumen     []StatusItem `json:"dokumen"`
	// QRToken is shown as a QR code for manasik attendance (haji only).
	QRToken string `json:"qr_token,omitempty"`
}

// CekKamar godoc
// @Summary      Cek kamar manasik (publik)
// @Description  Jamaah memasukkan NIK dan tanggal lahir untuk melihat gedung dan kamar manasik di semua paket yang diikuti
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

	ctx := c.Request.Context()
	list, err := h.repo.FindAllByNIK(ctx, strings.TrimSpace(req.NIK))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "terjadi kesalahan, coba lagi nanti"})
		return
	}

	resp := CekKamarResponse{Paket: []CekKamarPaket{}}
	for i := range list {
		jamaah := &list[i]
		// Each registration must match the birth date on its own.
		if jamaah.TanggalLahir.IsZero() || jamaah.TanggalLahir.UTC().Format("2006-01-02") != lahir.Format("2006-01-02") {
			continue
		}
		// Newest registration's name wins (most likely the corrected one).
		resp.Nama = jamaah.Nama
		resp.Paket = append(resp.Paket, h.cekKamarPaket(ctx, jamaah))
	}
	if len(resp.Paket) == 0 {
		c.JSON(http.StatusNotFound, notFound)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// kelengkapan lists the items a jamaah receives: haji gets batik, umroh a
// seragam; both get a koper.
func kelengkapan(j *model.Jamaah, tipe string) []StatusItem {
	koper := StatusItem{Nama: "Koper sudah diterima", Selesai: j.KoperSudahDiterima}
	switch tipe {
	case "haji":
		return []StatusItem{
			{Nama: "Batik Nasional sudah dijahit", Selesai: j.BatikNasionalSudahDijahit},
			{Nama: "Batik KBIH sudah diterima", Selesai: j.BatikKBIHSudahDiterima},
			koper,
		}
	case "umroh":
		return []StatusItem{{Nama: "Seragam umroh sudah diterima", Selesai: j.SeragamUmrohSudahDiterima}, koper}
	}
	return []StatusItem{koper}
}

func (h *PortalHandler) cekKamarPaket(ctx context.Context, jamaah *model.Jamaah) CekKamarPaket {
	out := CekKamarPaket{
		TanggalKeberangkatan: jamaah.TanggalKeberangkatan,
		GedungManasik:        jamaah.GedungManasik,
		KamarManasik:         jamaah.KamarManasik,
		Rombongan:            jamaah.Rombongan,
		Regu:                 jamaah.Regu,
		KetuaRegu:            []KontakKetua{},
		KetuaRombongan:       []KontakKetua{},
		Dokumen: []StatusItem{
			{Nama: "KTP", Selesai: jamaah.FotoKTP != ""},
			{Nama: "Kartu Keluarga", Selesai: jamaah.FotoKK != ""},
			{Nama: "Paspor", Selesai: jamaah.FotoPaspor != ""},
			{Nama: "Pas foto", Selesai: jamaah.Pasfoto != ""},
		},
	}
	if jamaah.PaketID.IsZero() {
		out.Kelengkapan = kelengkapan(jamaah, "")
		return out
	}
	paket, err := h.paketRepo.FindByID(ctx, jamaah.PaketID)
	if err != nil {
		out.Kelengkapan = kelengkapan(jamaah, "")
		return out
	}
	out.Paket = paket.Label
	out.TipePaket = paket.Tipe
	out.Kelengkapan = kelengkapan(jamaah, paket.Tipe)
	out.HotelMekkah, out.HotelMadinah = paket.HotelMekkah, paket.HotelMadinah
	out.KamarHotelMekkah, out.KamarHotelMadinah = jamaah.KamarHotelMekkah, jamaah.KamarHotelMadinah
	out.MapsHotelMekkah, out.MapsHotelMadinah = paket.MapsHotelMekkah, paket.MapsHotelMadinah
	if paket.Tipe == "haji" {
		out.QRToken = manasikToken(h.qrSecret, jamaah.ID)
		if regu, rombongan, err := h.repo.FindLeaders(ctx, jamaah.PaketID, jamaah.Rombongan, jamaah.Regu); err == nil {
			out.KetuaRegu = toKontak(regu)
			out.KetuaRombongan = toKontak(rombongan)
		}
	}
	return out
}

func (h *PortalHandler) RegisterRoutes(r *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	handlers := append(middlewares, h.CekKamar)
	r.POST("/portal/cek-kamar", handlers...)
}
