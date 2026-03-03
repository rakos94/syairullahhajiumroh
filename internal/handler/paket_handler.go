package handler

import (
	"net/http"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaketHandler struct {
	repo *repository.PaketRepository
}

func NewPaketHandler(repo *repository.PaketRepository) *PaketHandler {
	return &PaketHandler{repo: repo}
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

	if err := h.repo.Create(c.Request.Context(), &paket); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "paket sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	paket.BuildLabel()
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

	if err := h.repo.Update(c.Request.Context(), id, &paket); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "paket sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "paket berhasil dihapus"})
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
