package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var validDocTypes = map[string]string{
	"ktp":             "foto_ktp",
	"kk":              "foto_kk",
	"paspor":          "foto_paspor",
	"pasfoto":         "pasfoto",
	"koper_diterima":  "foto_koper_diterima",
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

type JamaahHandler struct {
	repo      *repository.JamaahRepository
	uploadDir string
}

func NewJamaahHandler(repo *repository.JamaahRepository, uploadDir string) *JamaahHandler {
	return &JamaahHandler{repo: repo, uploadDir: uploadDir}
}

// Create godoc
// @Summary      Tambah jamaah baru
// @Description  Membuat data jamaah haji/umroh baru
// @Tags         jamaah
// @Accept       json
// @Produce      json
// @Param        jamaah  body      model.Jamaah  true  "Data jamaah"
// @Success      201     {object}  model.Jamaah
// @Failure      400     {object}  map[string]string
// @Failure      409     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /jamaah [post]
func (h *JamaahHandler) Create(c *gin.Context) {
	var jamaah model.Jamaah
	if err := c.ShouldBindJSON(&jamaah); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(c.Request.Context(), &jamaah); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "jamaah dengan NIK atau nomor paspor tersebut sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, jamaah)
}

// FindAll godoc
// @Summary      List semua jamaah
// @Description  Menampilkan daftar jamaah, bisa difilter berdasarkan paket
// @Tags         jamaah
// @Produce      json
// @Param        paket  query     string  false  "Filter berdasarkan paket"  Enums(haji, umroh)
// @Success      200    {array}   model.Jamaah
// @Failure      500    {object}  map[string]string
// @Router       /jamaah [get]
func (h *JamaahHandler) FindAll(c *gin.Context) {
	paket := c.Query("paket")

	results, err := h.repo.FindAll(c.Request.Context(), paket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// FindByID godoc
// @Summary      Detail jamaah
// @Description  Menampilkan detail jamaah berdasarkan ID
// @Tags         jamaah
// @Produce      json
// @Param        id   path      string  true  "Jamaah ID"
// @Success      200  {object}  model.Jamaah
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /jamaah/{id} [get]
func (h *JamaahHandler) FindByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	jamaah, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jamaah)
}

// Update godoc
// @Summary      Update jamaah
// @Description  Memperbarui data jamaah berdasarkan ID
// @Tags         jamaah
// @Accept       json
// @Produce      json
// @Param        id      path      string        true  "Jamaah ID"
// @Param        jamaah  body      model.Jamaah  true  "Data jamaah"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      409     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /jamaah/{id} [put]
func (h *JamaahHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	var jamaah model.Jamaah
	if err := c.ShouldBindJSON(&jamaah); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(c.Request.Context(), id, &jamaah); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "jamaah dengan NIK atau nomor paspor tersebut sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "jamaah berhasil diperbarui"})
}

// Delete godoc
// @Summary      Hapus jamaah
// @Description  Menghapus data jamaah berdasarkan ID
// @Tags         jamaah
// @Produce      json
// @Param        id   path      string  true  "Jamaah ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /jamaah/{id} [delete]
func (h *JamaahHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "jamaah berhasil dihapus"})
}

// UploadDocument godoc
// @Summary      Upload dokumen jamaah
// @Description  Upload file dokumen (KTP, KK, Paspor, Pasfoto, Koper Diterima) untuk jamaah
// @Tags         dokumen
// @Accept       multipart/form-data
// @Produce      json
// @Param        id       path      string  true  "Jamaah ID"
// @Param        docType  path      string  true  "Tipe dokumen"  Enums(ktp, kk, paspor, pasfoto, koper_diterima)
// @Param        file     formData  file    true  "File dokumen (jpg, jpeg, png, pdf)"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /jamaah/{id}/upload/{docType} [post]
func (h *JamaahHandler) UploadDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	docType := c.Param("docType")
	bsonField, ok := validDocTypes[docType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tipe dokumen tidak valid, gunakan: %s", validDocTypesList())})
		return
	}

	// Check jamaah exists
	_, err = h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file tidak ditemukan"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format file tidak didukung, gunakan: jpg, jpeg, png, pdf"})
		return
	}

	// Create directory uploads/{jamaah_id}/
	dir := filepath.Join(h.uploadDir, id.Hex())
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat direktori"})
		return
	}

	filename := docType + ext
	filePath := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file"})
		return
	}

	// Update jamaah record with relative path
	relativePath := filepath.Join(id.Hex(), filename)
	if err := h.repo.UpdateField(c.Request.Context(), id, bsonField, relativePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("dokumen %s berhasil diupload", docType), "path": relativePath})
}

// GetDocument godoc
// @Summary      Download dokumen jamaah
// @Description  Download file dokumen jamaah berdasarkan tipe
// @Tags         dokumen
// @Produce      octet-stream
// @Param        id       path      string  true  "Jamaah ID"
// @Param        docType  path      string  true  "Tipe dokumen"  Enums(ktp, kk, paspor, pasfoto, koper_diterima)
// @Success      200      {file}    file
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /jamaah/{id}/dokumen/{docType} [get]
func (h *JamaahHandler) GetDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	docType := c.Param("docType")
	bsonField, ok := validDocTypes[docType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tipe dokumen tidak valid, gunakan: %s", validDocTypesList())})
		return
	}

	jamaah, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var relativePath string
	switch bsonField {
	case "foto_ktp":
		relativePath = jamaah.FotoKTP
	case "foto_kk":
		relativePath = jamaah.FotoKK
	case "foto_paspor":
		relativePath = jamaah.FotoPaspor
	case "pasfoto":
		relativePath = jamaah.Pasfoto
	case "foto_koper_diterima":
		relativePath = jamaah.FotoKoperDiterima
	}

	if relativePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("dokumen %s belum diupload", docType)})
		return
	}

	filePath := filepath.Join(h.uploadDir, relativePath)
	c.File(filePath)
}

func validDocTypesList() string {
	types := make([]string, 0, len(validDocTypes))
	for k := range validDocTypes {
		types = append(types, k)
	}
	return strings.Join(types, ", ")
}

func (h *JamaahHandler) RegisterRoutes(r *gin.RouterGroup) {
	jamaah := r.Group("/jamaah")
	{
		jamaah.POST("", h.Create)
		jamaah.GET("", h.FindAll)
		jamaah.GET("/:id", h.FindByID)
		jamaah.PUT("/:id", h.Update)
		jamaah.DELETE("/:id", h.Delete)
		jamaah.POST("/:id/upload/:docType", h.UploadDocument)
		jamaah.GET("/:id/dokumen/:docType", h.GetDocument)
	}
}
