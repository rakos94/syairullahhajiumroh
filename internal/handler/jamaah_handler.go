package handler

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	"bukti_dp":        "bukti_dp",
	"bukti_pelunasan": "bukti_pelunasan",
}

var multiDocTypes = map[string]bool{
	"bukti_dp":        true,
	"bukti_pelunasan": true,
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
}

type JamaahHandler struct {
	repo      *repository.JamaahRepository
	paketRepo *repository.PaketRepository
	uploadDir string
}

func NewJamaahHandler(repo *repository.JamaahRepository, paketRepo *repository.PaketRepository, uploadDir string) *JamaahHandler {
	return &JamaahHandler{repo: repo, paketRepo: paketRepo, uploadDir: uploadDir}
}

// populatePaket populates the transient Paket field for a slice of jamaah.
func (h *JamaahHandler) populatePaket(c *gin.Context, list []model.Jamaah) {
	ids := make(map[primitive.ObjectID]struct{})
	for _, j := range list {
		if !j.PaketID.IsZero() {
			ids[j.PaketID] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return
	}

	idSlice := make([]primitive.ObjectID, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}

	pakets, err := h.paketRepo.FindByIDs(c.Request.Context(), idSlice)
	if err != nil {
		return
	}

	paketMap := make(map[primitive.ObjectID]*model.Paket, len(pakets))
	for i := range pakets {
		paketMap[pakets[i].ID] = &pakets[i]
	}

	for i := range list {
		if p, ok := paketMap[list[i].PaketID]; ok {
			list[i].Paket = p
		}
	}
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

	// Validate paket_id exists
	paket, err := h.paketRepo.FindByID(c.Request.Context(), jamaah.PaketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paket_id tidak valid"})
		return
	}

	// Validate departure date for haji pakets
	if paket.Tipe == "haji" && len(paket.TanggalKeberangkatan) > 0 {
		if jamaah.TanggalKeberangkatan == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal keberangkatan harus dipilih"})
			return
		}
		found := false
		for _, tk := range paket.TanggalKeberangkatan {
			if tk.Nama == jamaah.TanggalKeberangkatan.Nama {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal keberangkatan tidak valid untuk paket ini"})
			return
		}
	}

	if err := h.repo.Create(c.Request.Context(), &jamaah); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "jamaah dengan NIK atau nomor paspor tersebut sudah ada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jamaah.Paket = paket
	c.JSON(http.StatusCreated, jamaah)
}

// FindAll godoc
// @Summary      List semua jamaah
// @Description  Menampilkan daftar jamaah dengan pagination, bisa difilter berdasarkan paket
// @Tags         jamaah
// @Produce      json
// @Param        paket_id  query     string  false  "Filter berdasarkan paket ID"
// @Param        page      query     int     false  "Halaman"       default(1)
// @Param        limit     query     int     false  "Jumlah per halaman"  default(10)
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /jamaah [get]
func (h *JamaahHandler) FindAll(c *gin.Context) {
	var paketID *primitive.ObjectID
	if pidStr := c.Query("paket_id"); pidStr != "" {
		if oid, err := primitive.ObjectIDFromHex(pidStr); err == nil {
			paketID = &oid
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	results, total, err := h.repo.FindAll(c.Request.Context(), paketID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.populatePaket(c, results)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data":        results,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
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

	// Populate paket
	if !jamaah.PaketID.IsZero() {
		if paket, err := h.paketRepo.FindByID(c.Request.Context(), jamaah.PaketID); err == nil {
			jamaah.Paket = paket
		}
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

	// Validate paket_id exists
	paket, err := h.paketRepo.FindByID(c.Request.Context(), jamaah.PaketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paket_id tidak valid"})
		return
	}

	// Validate departure date for haji pakets
	if paket.Tipe == "haji" && len(paket.TanggalKeberangkatan) > 0 {
		if jamaah.TanggalKeberangkatan == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal keberangkatan harus dipilih"})
			return
		}
		found := false
		for _, tk := range paket.TanggalKeberangkatan {
			if tk.Nama == jamaah.TanggalKeberangkatan.Nama {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tanggal keberangkatan tidak valid untuk paket ini"})
			return
		}
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
	jamaah, err := h.repo.FindByID(c.Request.Context(), id)
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

	// Delete old file for single-doc types
	if !multiDocTypes[docType] {
		var oldPath string
		switch bsonField {
		case "foto_ktp":
			oldPath = jamaah.FotoKTP
		case "foto_kk":
			oldPath = jamaah.FotoKK
		case "foto_paspor":
			oldPath = jamaah.FotoPaspor
		case "pasfoto":
			oldPath = jamaah.Pasfoto
		case "foto_koper_diterima":
			oldPath = jamaah.FotoKoperDiterima
		}
		if oldPath != "" {
			os.Remove(filepath.Join(h.uploadDir, oldPath))
		}
	}

	var filename string
	if multiDocTypes[docType] {
		filename = fmt.Sprintf("%s_%d%s", docType, time.Now().UnixMilli(), ext)
	} else {
		filename = docType + ext
	}
	filePath := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file"})
		return
	}

	// Update jamaah record with relative path
	relativePath := filepath.Join(id.Hex(), filename)
	if multiDocTypes[docType] {
		nominal, _ := strconv.ParseFloat(c.PostForm("nominal"), 64)
		entry := model.BuktiPembayaran{File: relativePath, Nominal: nominal}
		if err := h.repo.PushToArray(c.Request.Context(), id, bsonField, entry); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := h.repo.UpdateField(c.Request.Context(), id, bsonField, relativePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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

	if multiDocTypes[docType] {
		var entries []model.BuktiPembayaran
		switch bsonField {
		case "bukti_dp":
			entries = jamaah.BuktiDP
		case "bukti_pelunasan":
			entries = jamaah.BuktiPelunasan
		}
		idxStr := c.Query("index")
		idx, err := strconv.Atoi(idxStr)
		if err != nil || idx < 0 || idx >= len(entries) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("dokumen %s tidak ditemukan", docType)})
			return
		}
		c.File(filepath.Join(h.uploadDir, entries[idx].File))
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

func (h *JamaahHandler) DeleteDocument(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}

	docType := c.Param("docType")
	bsonField, ok := validDocTypes[docType]
	if !ok || !multiDocTypes[docType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipe dokumen tidak valid"})
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

	var entries []model.BuktiPembayaran
	switch bsonField {
	case "bukti_dp":
		entries = jamaah.BuktiDP
	case "bukti_pelunasan":
		entries = jamaah.BuktiPelunasan
	}

	idxStr := c.Query("index")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 || idx >= len(entries) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index tidak valid"})
		return
	}

	entry := entries[idx]

	// Remove file
	os.Remove(filepath.Join(h.uploadDir, entry.File))

	// Remove from array in DB
	if err := h.repo.PullFromArray(c.Request.Context(), id, bsonField, entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("dokumen %s berhasil dihapus", docType)})
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
		jamaah.DELETE("/:id/dokumen/:docType", h.DeleteDocument)
	}
}
