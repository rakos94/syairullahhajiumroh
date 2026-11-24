package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaketHandler struct {
	repo       *repository.PaketRepository
	jamaahRepo *repository.JamaahRepository
	auditRepo  *repository.AuditLogRepository
}

func NewPaketHandler(repo *repository.PaketRepository, jamaahRepo *repository.JamaahRepository, auditRepo *repository.AuditLogRepository) *PaketHandler {
	return &PaketHandler{repo: repo, jamaahRepo: jamaahRepo, auditRepo: auditRepo}
}

// CreatePaket godoc
// @Summary      Tambah paket baru
// @Description  Membuat data paket haji/umroh baru
// @Tags         paket
// @Accept       json
// @Produce      json
// @Param        paket  body      model.Paket  true  "Data paket"
// @Success      201    {object}  model.Paket
// @Failure      400    {object}  map[string]string
// @Failure      409    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /paket [post]
func (h *PaketHandler) Create(c *gin.Context) {
	var paket model.Paket
	if err := c.ShouldBindJSON(&paket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if paket.Tipe == "umroh" && (paket.Bulan < 1 || paket.Bulan > 12) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bulan harus 1-12 untuk paket umroh"})
		return
	}
	if paket.Tipe == "haji" {
		paket.Bulan = 0
	}
	if msg := normalizeHotel(&paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if paket.Harga < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "harga tidak boleh negatif"})
		return
	}
	if msg := normalizePilihanHarga(&paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	for i, tk := range paket.TanggalKeberangkatan {
		if strings.TrimSpace(tk.Nama) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tanggal keberangkatan ke-%d: nama harus diisi", i+1)})
			return
		}
	}

	if err := h.repo.Create(c.Request.Context(), &paket); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "paket sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paket.BuildLabel()
	logAudit(c, h.auditRepo, "paket", paket.ID, paket.Label, "create", fmt.Sprintf("Menambah paket: %s", paket.Label))

	c.JSON(http.StatusCreated, paket)
}

// FindAllPaket godoc
// @Summary      List semua paket
// @Description  Menampilkan daftar paket, bisa difilter berdasarkan tipe
// @Tags         paket
// @Produce      json
// @Param        tipe  query     string  false  "Filter berdasarkan tipe"  Enums(haji, umroh)
// @Success      200   {array}   model.Paket
// @Failure      500   {object}  map[string]string
// @Router       /paket [get]
func (h *PaketHandler) FindAll(c *gin.Context) {
	tipe := c.Query("tipe")

	results, err := h.repo.FindAll(c.Request.Context(), tipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// FindByIDPaket godoc
// @Summary      Detail paket
// @Description  Menampilkan detail paket berdasarkan ID
// @Tags         paket
// @Produce      json
// @Param        id   path      string  true  "Paket ID"
// @Success      200  {object}  model.Paket
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /paket/{id} [get]
func (h *PaketHandler) FindByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	paket, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "paket tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paket)
}

// UpdatePaket godoc
// @Summary      Update paket
// @Description  Memperbarui data paket berdasarkan ID
// @Tags         paket
// @Accept       json
// @Produce      json
// @Param        id     path      string       true  "Paket ID"
// @Param        paket  body      model.Paket  true  "Data paket"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      409    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /paket/{id} [put]
func (h *PaketHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	// Fetch old data for diff
	oldPaket, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "paket tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var paket model.Paket
	if err := c.ShouldBindJSON(&paket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if paket.Tipe == "umroh" && (paket.Bulan < 1 || paket.Bulan > 12) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bulan harus 1-12 untuk paket umroh"})
		return
	}
	if paket.Tipe == "haji" {
		paket.Bulan = 0
	}
	if msg := normalizeHotel(&paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if paket.Harga < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "harga tidak boleh negatif"})
		return
	}
	if msg := normalizePilihanHarga(&paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	for i, tk := range paket.TanggalKeberangkatan {
		if strings.TrimSpace(tk.Nama) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tanggal keberangkatan ke-%d: nama harus diisi", i+1)})
			return
		}
	}

	// Don't silently change the price of jamaah who use an option being removed/renamed.
	var removed []string
	for _, opt := range oldPaket.PilihanHarga {
		if _, ok := paket.FindPilihanHarga(opt.Nama); !ok {
			removed = append(removed, opt.Nama)
		}
	}
	if len(removed) > 0 {
		inUse, err := h.jamaahRepo.CountUsingPilihanHarga(c.Request.Context(), id, removed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var msgs []string
		for _, name := range removed {
			if n := inUse[name]; n > 0 {
				msgs = append(msgs, fmt.Sprintf("%q dipakai %d jamaah", name, n))
			}
		}
		if len(msgs) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "pilihan harga tidak bisa dihapus atau diganti namanya karena masih dipakai: " +
				strings.Join(msgs, ", ") + ". Pindahkan jamaah tersebut ke pilihan lain terlebih dahulu."})
			return
		}
	}

	if err := h.repo.Update(c.Request.Context(), id, &paket); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "paket sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Price may have changed: recompute payment status of its jamaah.
	if updated, err := h.repo.FindByID(c.Request.Context(), id); err == nil {
		if list, err := h.jamaahRepo.FindActive(c.Request.Context(), &id); err == nil {
			for i := range list {
				syncStatusPembayaran(c, h.jamaahRepo, h.auditRepo, &list[i], updated)
			}
		}
	}

	// Sync jamaah departure dates
	for _, tk := range paket.TanggalKeberangkatan {
		_ = h.jamaahRepo.UpdateDepartureByPaket(c.Request.Context(), id, tk.Nama, tk.Tanggal)
	}

	paket.BuildLabel()
	changes := diffPaket(oldPaket, &paket)
	logAuditWithChanges(c, h.auditRepo, "paket", id, paket.Label, "update", fmt.Sprintf("Memperbarui paket: %s", paket.Label), changes)

	c.JSON(http.StatusOK, gin.H{"message": "paket berhasil diperbarui"})
}

// DeletePaket godoc
// @Summary      Hapus paket
// @Description  Menghapus data paket berdasarkan ID
// @Tags         paket
// @Produce      json
// @Param        id   path      string  true  "Paket ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /paket/{id} [delete]
func (h *PaketHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	paket, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "paket tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logAudit(c, h.auditRepo, "paket", id, paket.Label, "delete", fmt.Sprintf("Menghapus paket: %s", paket.Label))

	c.JSON(http.StatusOK, gin.H{"message": "paket berhasil dihapus"})
}

// normalizePilihanHarga trims names and validates price options.
// normalizeHotel trims hotel names and checks the Maps links; hotels are
// only for umroh pakets. It returns an error message, or "" when valid.
func normalizeHotel(p *model.Paket) string {
	p.HotelMekkah = strings.TrimSpace(p.HotelMekkah)
	p.HotelMadinah = strings.TrimSpace(p.HotelMadinah)
	p.MapsHotelMekkah = strings.TrimSpace(p.MapsHotelMekkah)
	p.MapsHotelMadinah = strings.TrimSpace(p.MapsHotelMadinah)
	if p.Tipe != "umroh" {
		p.HotelMekkah, p.HotelMadinah, p.MapsHotelMekkah, p.MapsHotelMadinah = "", "", "", ""
		return ""
	}
	if p.MapsHotelMekkah != "" && !isGoogleMapsURL(p.MapsHotelMekkah) {
		return "link Google Maps hotel Mekkah tidak valid (salin dari tombol Bagikan di Google Maps)"
	}
	if p.MapsHotelMadinah != "" && !isGoogleMapsURL(p.MapsHotelMadinah) {
		return "link Google Maps hotel Madinah tidak valid (salin dari tombol Bagikan di Google Maps)"
	}
	return ""
}

// isGoogleMapsURL accepts only https links to Google Maps; these links are
// shown on the public Cek Kamar page.
func isGoogleMapsURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "maps.app.goo.gl", host == "maps.google.com", strings.HasPrefix(host, "maps.google."):
		return true
	case host == "goo.gl":
		return strings.HasPrefix(u.Path, "/maps")
	case host == "google.com", host == "www.google.com", strings.HasPrefix(host, "www.google."):
		return strings.HasPrefix(u.Path, "/maps")
	}
	return false
}

func normalizePilihanHarga(p *model.Paket) string {
	seen := map[string]bool{}
	for i := range p.PilihanHarga {
		opt := &p.PilihanHarga[i]
		opt.Nama = strings.TrimSpace(opt.Nama)
		if opt.Nama == "" {
			return fmt.Sprintf("pilihan harga ke-%d: nama harus diisi", i+1)
		}
		key := strings.ToLower(opt.Nama)
		if seen[key] {
			return fmt.Sprintf("pilihan harga %q duplikat", opt.Nama)
		}
		seen[key] = true
		if opt.Harga <= 0 {
			return fmt.Sprintf("pilihan harga %q: harga harus lebih dari 0", opt.Nama)
		}
	}
	return ""
}

func (h *PaketHandler) RegisterRoutes(r *gin.RouterGroup) {
	paket := r.Group("/paket")
	{
		paket.POST("", h.Create)
		paket.GET("", h.FindAll)
		paket.GET("/:id", h.FindByID)
		paket.PUT("/:id", h.Update)
		paket.DELETE("/:id", h.Delete)
	}
}
