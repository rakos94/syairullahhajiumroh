package handler

import (
	"fmt"
	"net/http"
	"strings"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const publicPengumumanLimit = 20

type PengumumanHandler struct {
	repo      *repository.PengumumanRepository
	auditRepo *repository.AuditLogRepository
}

func NewPengumumanHandler(repo *repository.PengumumanRepository, auditRepo *repository.AuditLogRepository) *PengumumanHandler {
	return &PengumumanHandler{repo: repo, auditRepo: auditRepo}
}

func normalizePengumuman(p *model.Pengumuman) error {
	p.Judul = strings.TrimSpace(p.Judul)
	p.Isi = strings.TrimSpace(p.Isi)
	if p.Judul == "" || p.Isi == "" {
		return fmt.Errorf("judul dan isi pengumuman wajib diisi")
	}
	if len(p.Judul) > 200 {
		return fmt.Errorf("judul maksimal 200 karakter")
	}
	if len(p.Isi) > 5000 {
		return fmt.Errorf("isi maksimal 5000 karakter")
	}
	return nil
}

// ListPublic godoc
// @Summary      Daftar pengumuman (publik)
// @Description  Pengumuman yang dipublikasikan, yang disematkan di atas lalu terbaru
// @Tags         portal
// @Produce      json
// @Success      200  {array}  model.Pengumuman
// @Router       /portal/pengumuman [get]
func (h *PengumumanHandler) ListPublic(c *gin.Context) {
	list, err := h.repo.FindAll(c.Request.Context(), true, publicPengumumanLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memuat pengumuman"})
		return
	}
	// Don't expose which admin wrote it on the public endpoint.
	for i := range list {
		list[i].CreatedBy = ""
	}
	c.JSON(http.StatusOK, list)
}

// List godoc
// @Summary      Daftar semua pengumuman (admin)
// @Tags         pengumuman
// @Produce      json
// @Success      200  {array}  model.Pengumuman
// @Router       /pengumuman [get]
func (h *PengumumanHandler) List(c *gin.Context) {
	list, err := h.repo.FindAll(c.Request.Context(), false, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Create godoc
// @Summary      Tambah pengumuman
// @Tags         pengumuman
// @Accept       json
// @Produce      json
// @Param        body  body      model.Pengumuman  true  "Pengumuman"
// @Success      201   {object}  model.Pengumuman
// @Router       /pengumuman [post]
func (h *PengumumanHandler) Create(c *gin.Context) {
	var p model.Pengumuman
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "judul dan isi pengumuman wajib diisi"})
		return
	}
	if err := normalizePengumuman(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if u, ok := c.Get("admin_username"); ok {
		p.CreatedBy, _ = u.(string)
	}
	if err := h.repo.Create(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logAudit(c, h.auditRepo, "pengumuman", p.ID, p.Judul, "create", fmt.Sprintf("Menambah pengumuman: %s", p.Judul))
	c.JSON(http.StatusCreated, p)
}

// Update godoc
// @Summary      Ubah pengumuman
// @Tags         pengumuman
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "ID pengumuman"
// @Param        body  body      model.Pengumuman  true  "Pengumuman"
// @Success      200   {object}  model.Pengumuman
// @Router       /pengumuman/{id} [put]
func (h *PengumumanHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	old, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "pengumuman tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var p model.Pengumuman
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "judul dan isi pengumuman wajib diisi"})
		return
	}
	if err := normalizePengumuman(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Update(c.Request.Context(), id, &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var changes []model.FieldChange
	diffField(&changes, "Judul", old.Judul, p.Judul)
	diffField(&changes, "Isi", old.Isi, p.Isi)
	diffField(&changes, "Dipublikasikan", formatBool(old.Dipublikasikan), formatBool(p.Dipublikasikan))
	diffField(&changes, "Disematkan", formatBool(old.Disematkan), formatBool(p.Disematkan))
	logAuditWithChanges(c, h.auditRepo, "pengumuman", id, p.Judul, "update", fmt.Sprintf("Memperbarui pengumuman: %s", p.Judul), changes)

	updated, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "pengumuman berhasil diperbarui"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete godoc
// @Summary      Hapus pengumuman
// @Tags         pengumuman
// @Produce      json
// @Param        id  path  string  true  "ID pengumuman"
// @Success      200  {object}  map[string]string
// @Router       /pengumuman/{id} [delete]
func (h *PengumumanHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	old, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "pengumuman tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logAudit(c, h.auditRepo, "pengumuman", id, old.Judul, "delete", fmt.Sprintf("Menghapus pengumuman: %s", old.Judul))
	c.JSON(http.StatusOK, gin.H{"message": "pengumuman berhasil dihapus"})
}

func (h *PengumumanHandler) RegisterPublicRoutes(r *gin.RouterGroup) {
	r.GET("/portal/pengumuman", h.ListPublic)
}

func (h *PengumumanHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/pengumuman")
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
