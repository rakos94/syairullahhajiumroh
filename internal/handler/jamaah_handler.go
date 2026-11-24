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

	"syairullahhajiumroh/internal/middleware"
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
	auditRepo *repository.AuditLogRepository
	uploadDir string
	secret    string // signs document links and checks admin tokens
}

func NewJamaahHandler(repo *repository.JamaahRepository, paketRepo *repository.PaketRepository, auditRepo *repository.AuditLogRepository, uploadDir, secret string) *JamaahHandler {
	return &JamaahHandler{repo: repo, paketRepo: paketRepo, auditRepo: auditRepo, uploadDir: uploadDir, secret: secret}
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
		list[i].Tagihan = model.HitungTagihan(&list[i], list[i].Paket)
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
	if msg := validateHargaJamaah(&jamaah, paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Validate departure date for pakets with departure options
	if len(paket.TanggalKeberangkatan) > 0 {
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

	// Documents are only set through the upload endpoints. Identity
	// documents of a person already registered in another paket are reused.
	jamaah.FotoKTP, jamaah.FotoKK, jamaah.FotoPaspor, jamaah.Pasfoto, jamaah.FotoKoperDiterima = "", "", "", "", ""
	jamaah.BuktiDP, jamaah.BuktiPelunasan = nil, nil
	others, err := h.repo.FindAllByNIK(c.Request.Context(), jamaah.NIK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(others) > 0 {
		// The person is already registered in another paket. A different
		// birth date most likely means a mistyped NIK; refuse rather than
		// overwrite someone else's data.
		last := &others[len(others)-1]
		if !sameDay(last.TanggalLahir, jamaah.TanggalLahir) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("NIK ini sudah terdaftar atas nama %s dengan tanggal lahir berbeda (%s). Periksa kembali NIK, atau perbaiki tanggal lahir di data yang sudah ada.", last.Nama, formatDate(last.TanggalLahir))})
			return
		}
		if msg, err := h.konflikSinkron(c, primitive.NilObjectID, others, &jamaah); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if msg != "" {
			c.JSON(http.StatusConflict, gin.H{"error": msg})
			return
		}
		// Newest registration first, so the latest upload wins.
		for i := len(others) - 1; i >= 0; i-- {
			o := &others[i]
			for _, d := range []struct {
				dst *string
				src string
			}{
				{&jamaah.FotoKTP, o.FotoKTP}, {&jamaah.FotoKK, o.FotoKK},
				{&jamaah.FotoPaspor, o.FotoPaspor}, {&jamaah.Pasfoto, o.Pasfoto},
			} {
				if *d.dst == "" {
					*d.dst = d.src
				}
			}
		}
	}

	applyStatusOtomatis(&jamaah, paket)

	if err := h.repo.Create(c.Request.Context(), &jamaah); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "NIK atau nomor paspor tersebut sudah terdaftar di paket ini"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logAudit(c, h.auditRepo, "jamaah", jamaah.ID, jamaah.Nama, "create", fmt.Sprintf("Menambah jamaah: %s", jamaah.Nama))

	// What was entered now is the person's latest data diri.
	if err := h.sinkronDataDiri(c, others, &jamaah, paket.Label); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "data tersimpan, tetapi gagal memperbarui data diri di paket lain: " + err.Error()})
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

	search := c.Query("search")
	statusPembayaran := c.Query("status_pembayaran")
	kelengkapan := c.Query("kelengkapan")

	results, total, err := h.repo.FindAll(c.Request.Context(), paketID, search, statusPembayaran, kelengkapan, page, limit)
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
	jamaah.Tagihan = model.HitungTagihan(jamaah, jamaah.Paket)
	jamaah.DokumenAkses = newDokumenAkses(h.secret, jamaah.ID)

	// Same person registered in other pakets.
	if others, err := h.repo.FindAllByNIK(c.Request.Context(), jamaah.NIK); err == nil {
		for _, o := range others {
			if o.ID == jamaah.ID {
				continue
			}
			pl := model.PaketLain{JamaahID: o.ID, PaketID: o.PaketID, StatusPembayaran: o.StatusPembayaran}
			if p, err := h.paketRepo.FindByID(c.Request.Context(), o.PaketID); err == nil {
				pl.Paket = p.Label
			}
			jamaah.PaketLain = append(jamaah.PaketLain, pl)
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

	// Fetch old data for diff
	oldJamaah, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if msg := validateHargaJamaah(&jamaah, paket); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Validate departure date for pakets with departure options
	if len(paket.TanggalKeberangkatan) > 0 {
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

	// The edit form never sends uploaded documents or creation time, and the
	// repository $sets the whole struct. Keep the stored values so an edit
	// doesn't wipe them (documents are managed via the upload endpoints).
	jamaah.FotoKTP = oldJamaah.FotoKTP
	jamaah.FotoKK = oldJamaah.FotoKK
	jamaah.FotoPaspor = oldJamaah.FotoPaspor
	jamaah.Pasfoto = oldJamaah.Pasfoto
	jamaah.FotoKoperDiterima = oldJamaah.FotoKoperDiterima
	jamaah.BuktiDP = oldJamaah.BuktiDP
	jamaah.BuktiPelunasan = oldJamaah.BuktiPelunasan
	jamaah.CreatedAt = oldJamaah.CreatedAt
	jamaah.DeletedAt = oldJamaah.DeletedAt
	// Receipts are known now, so the status can be derived (priced paket only).
	applyStatusOtomatis(&jamaah, paket)

	// The same person's other registrations get the new personal data too.
	// Check first that a changed NIK/passport doesn't collide in their
	// pakets, so the sync can't stop halfway.
	others, err := h.otherRegistrations(c, oldJamaah)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(others) > 0 && (jamaah.NIK != oldJamaah.NIK || jamaah.NomorPaspor != oldJamaah.NomorPaspor) {
		if msg, err := h.konflikSinkron(c, id, others, &jamaah); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if msg != "" {
			c.JSON(http.StatusConflict, gin.H{"error": msg})
			return
		}
	}

	if err := h.repo.Update(c.Request.Context(), id, &jamaah); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "NIK atau nomor paspor tersebut sudah terdaftar di paket ini"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	changes := diffJamaah(oldJamaah, &jamaah)
	logAuditWithChanges(c, h.auditRepo, "jamaah", id, jamaah.Nama, "update", fmt.Sprintf("Memperbarui jamaah: %s", jamaah.Nama), changes)

	if err := h.sinkronDataDiri(c, others, &jamaah, paket.Label); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "data tersimpan, tetapi gagal memperbarui data diri di paket lain: " + err.Error()})
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

	jamaah, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logAudit(c, h.auditRepo, "jamaah", id, jamaah.Nama, "delete", fmt.Sprintf("Menghapus jamaah: %s", jamaah.Nama))

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

	shared := isIdentityDoc(bsonField)

	// Delete old file for per-paket single-doc types. Shared identity
	// documents are cleaned up after the new path is saved everywhere.
	if !multiDocTypes[docType] && !shared {
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
	if multiDocTypes[docType] || shared {
		// Unique names: shared files may live in another registration's
		// folder and must never be overwritten in place.
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
	} else if shared {
		if err := h.saveSharedDoc(c, jamaah, docType, bsonField, relativePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := h.repo.UpdateField(c.Request.Context(), id, bsonField, relativePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	logAudit(c, h.auditRepo, "jamaah", id, jamaah.Nama, "upload", fmt.Sprintf("Upload dokumen %s untuk jamaah: %s", docType, jamaah.Nama))
	if multiDocTypes[docType] {
		h.syncStatusAfterReceiptChange(c, id)
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

	// Documents (KTP, KK, paspor, receipts...) are private: require either a
	// signed link from the admin detail page or an admin token.
	if !validDokumenAkses(h.secret, id, c.Query("exp"), c.Query("sig")) &&
		!middleware.ValidBearerToken(h.secret, c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "akses dokumen tidak valid atau sudah kedaluwarsa, muat ulang halaman"})
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

	var filePath string
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
		filePath = filepath.Join(h.uploadDir, entries[idx].File)
	} else {
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
		filePath = filepath.Join(h.uploadDir, relativePath)
	}

	ext := filepath.Ext(filePath)
	downloadName := fmt.Sprintf("%s_%s%s", jamaah.NIK, docType, ext)
	if c.Query("download") == "true" {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadName))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", downloadName))
	}
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

	logAudit(c, h.auditRepo, "jamaah", id, jamaah.Nama, "delete_document", fmt.Sprintf("Menghapus dokumen %s untuk jamaah: %s", docType, jamaah.Nama))
	h.syncStatusAfterReceiptChange(c, id)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("dokumen %s berhasil dihapus", docType)})
}

// otherRegistrations returns the active registrations of the same person
// (same NIK) in other pakets.
func (h *JamaahHandler) otherRegistrations(c *gin.Context, j *model.Jamaah) ([]model.Jamaah, error) {
	all, err := h.repo.FindAllByNIK(c.Request.Context(), j.NIK)
	if err != nil {
		return nil, err
	}
	out := make([]model.Jamaah, 0, len(all))
	for _, o := range all {
		if o.ID != j.ID {
			out = append(out, o)
		}
	}
	return out, nil
}

func sameDay(a, b time.Time) bool {
	return a.UTC().Format("2006-01-02") == b.UTC().Format("2006-01-02")
}

// konflikSinkron checks that j's NIK and passport can be synced to the
// person's other registrations without hitting the unique index in their
// pakets (so a sync can't stop halfway). It returns a message when not.
func (h *JamaahHandler) konflikSinkron(c *gin.Context, self primitive.ObjectID, others []model.Jamaah, j *model.Jamaah) (string, error) {
	group := []primitive.ObjectID{self}
	pakets := []primitive.ObjectID{}
	for _, o := range others {
		group = append(group, o.ID)
		pakets = append(pakets, o.PaketID)
	}
	konflik, err := h.repo.FindKonflikDataDiri(c.Request.Context(), group, pakets, j.NIK, j.NomorPaspor)
	if err != nil || konflik == nil {
		return "", err
	}
	label := "paket lain"
	if p, err := h.paketRepo.FindByID(c.Request.Context(), konflik.PaketID); err == nil {
		label = p.Label
	}
	return fmt.Sprintf("NIK atau nomor paspor tersebut sudah dipakai %s di %s, tempat jamaah ini juga terdaftar", konflik.Nama, label), nil
}

// sinkronDataDiri copies j's data diri to the person's other registrations
// and records the change in each one's Riwayat.
func (h *JamaahHandler) sinkronDataDiri(c *gin.Context, others []model.Jamaah, j *model.Jamaah, dari string) error {
	ids := []primitive.ObjectID{}
	for i := range others {
		if len(diffDataDiri(&others[i], j)) > 0 {
			ids = append(ids, others[i].ID)
		}
	}
	if err := h.repo.SyncDataDiri(c.Request.Context(), ids, j); err != nil {
		return err
	}
	for i := range others {
		o := &others[i]
		if changes := diffDataDiri(o, j); len(changes) > 0 {
			logAuditWithChanges(c, h.auditRepo, "jamaah", o.ID, j.Nama, "update",
				fmt.Sprintf("Data diri disinkronkan dari %s: %s", dari, j.Nama), changes)
		}
	}
	return nil
}

func isIdentityDoc(field string) bool {
	for _, f := range repository.IdentityDocFields {
		if f == field {
			return true
		}
	}
	return false
}

func docPath(j *model.Jamaah, field string) string {
	switch field {
	case "foto_ktp":
		return j.FotoKTP
	case "foto_kk":
		return j.FotoKK
	case "foto_paspor":
		return j.FotoPaspor
	case "pasfoto":
		return j.Pasfoto
	case "foto_koper_diterima":
		return j.FotoKoperDiterima
	}
	return ""
}

// saveSharedDoc stores a new identity document on every registration of
// the person, then deletes replaced files that no record points at anymore
// (soft-deleted records count, so their files are kept).
func (h *JamaahHandler) saveSharedDoc(c *gin.Context, j *model.Jamaah, docType, field, path string) error {
	ctx := c.Request.Context()
	others, err := h.otherRegistrations(c, j)
	if err != nil {
		return err
	}
	ids := []primitive.ObjectID{j.ID}
	oldPaths := map[string]bool{docPath(j, field): true}
	for i := range others {
		ids = append(ids, others[i].ID)
		oldPaths[docPath(&others[i], field)] = true
	}
	if err := h.repo.SetDocByIDs(ctx, ids, field, path); err != nil {
		return err
	}
	for i := range others {
		logAudit(c, h.auditRepo, "jamaah", others[i].ID, others[i].Nama, "upload",
			fmt.Sprintf("Dokumen %s disinkronkan dari paket lain: %s", docType, others[i].Nama))
	}
	for old := range oldPaths {
		if old == "" || old == path {
			continue
		}
		if n, err := h.repo.CountFileRefs(ctx, old); err == nil && n == 0 {
			os.Remove(filepath.Join(h.uploadDir, old))
		}
	}
	return nil
}

// syncStatusAfterReceiptChange reloads the jamaah (with its new receipts) and
// updates the payment status if the paket has a price.
func (h *JamaahHandler) syncStatusAfterReceiptChange(c *gin.Context, id primitive.ObjectID) {
	j, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		return
	}
	p, err := h.paketRepo.FindByID(c.Request.Context(), j.PaketID)
	if err != nil {
		return
	}
	syncStatusPembayaran(c, h.repo, h.auditRepo, j, p)
}

// validateHargaJamaah checks the chosen price option and discount against
// the paket. It returns an error message, or "" when valid.
func validateHargaJamaah(j *model.Jamaah, paket *model.Paket) string {
	j.PilihanHarga = strings.TrimSpace(j.PilihanHarga)
	j.KeteranganPotongan = strings.TrimSpace(j.KeteranganPotongan)
	if j.PilihanHarga != "" {
		if _, ok := paket.FindPilihanHarga(j.PilihanHarga); !ok {
			return fmt.Sprintf("pilihan harga %q tidak ada di %s", j.PilihanHarga, paket.Label)
		}
	}
	if j.Potongan < 0 {
		return "potongan tidak boleh negatif"
	}
	if dasar, _ := model.HargaDasarJamaah(j, paket); j.Potongan > 0 && dasar > 0 && j.Potongan > dasar {
		return fmt.Sprintf("potongan (%s) melebihi harga (%s)", formatRupiah(j.Potongan), formatRupiah(dasar))
	}
	if j.Potongan == 0 {
		j.KeteranganPotongan = ""
	}
	return ""
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
		jamaah.PUT("/bulk-manasik", h.BulkManasik)
		jamaah.GET("", h.FindAll)
		jamaah.GET("/:id", h.FindByID)
		jamaah.PUT("/:id", h.Update)
		jamaah.DELETE("/:id", h.Delete)
		jamaah.POST("/:id/upload/:docType", h.UploadDocument)
		jamaah.DELETE("/:id/dokumen/:docType", h.DeleteDocument)
	}
}
