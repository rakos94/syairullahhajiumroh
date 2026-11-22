package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	maxImportBytes = 5 << 20
	maxImportRows  = 3000
)

type ManasikHandler struct {
	repo       *repository.ManasikRepository
	jamaahRepo *repository.JamaahRepository
	paketRepo  *repository.PaketRepository
	auditRepo  *repository.AuditLogRepository
	qrSecret   string
}

func NewManasikHandler(repo *repository.ManasikRepository, jamaahRepo *repository.JamaahRepository, paketRepo *repository.PaketRepository, auditRepo *repository.AuditLogRepository, qrSecret string) *ManasikHandler {
	return &ManasikHandler{repo: repo, jamaahRepo: jamaahRepo, paketRepo: paketRepo, auditRepo: auditRepo, qrSecret: qrSecret}
}

// ---------- helpers ----------

// hajiPaket loads the paket from a hex id and checks it is a haji paket.
func (h *ManasikHandler) hajiPaket(c *gin.Context, idHex string) (*model.Paket, bool) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paket_id tidak valid"})
		return nil, false
	}
	paket, err := h.paketRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "paket tidak ditemukan"})
		return nil, false
	}
	if paket.Tipe != "haji" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manasik hanya untuk paket haji"})
		return nil, false
	}
	return paket, true
}

func adminUsername(c *gin.Context) string {
	if u, ok := c.Get("admin_username"); ok {
		if s, ok := u.(string); ok {
			return s
		}
	}
	return ""
}

// naturalLess compares strings so that "Kamar 2" < "Kamar 10".
// Empty strings sort last.
func naturalLess(a, b string) bool {
	if a == "" || b == "" {
		return a != "" && b == ""
	}
	ar, br := []rune(strings.ToLower(a)), []rune(strings.ToLower(b))
	i, j := 0, 0
	for i < len(ar) && j < len(br) {
		if unicode.IsDigit(ar[i]) && unicode.IsDigit(br[j]) {
			si := i
			for i < len(ar) && unicode.IsDigit(ar[i]) {
				i++
			}
			sj := j
			for j < len(br) && unicode.IsDigit(br[j]) {
				j++
			}
			na, _ := strconv.Atoi(string(ar[si:i]))
			nb, _ := strconv.Atoi(string(br[sj:j]))
			if na != nb {
				return na < nb
			}
			continue
		}
		if ar[i] != br[j] {
			return ar[i] < br[j]
		}
		i++
		j++
	}
	return len(ar)-i < len(br)-j
}

func sortByKamar(list []model.Jamaah) {
	sort.SliceStable(list, func(x, y int) bool {
		a, b := list[x], list[y]
		if a.GedungManasik != b.GedungManasik {
			return naturalLess(a.GedungManasik, b.GedungManasik)
		}
		if a.KamarManasik != b.KamarManasik {
			return naturalLess(a.KamarManasik, b.KamarManasik)
		}
		return a.Nama < b.Nama
	})
}

// ---------- rooming list ----------

// ListJamaah godoc
// @Summary      Daftar jamaah manasik per paket haji
// @Tags         manasik
// @Produce      json
// @Param        paket_id  query  string  true  "ID paket haji"
// @Success      200  {object}  map[string]interface{}
// @Router       /manasik/jamaah [get]
func (h *ManasikHandler) ListJamaah(c *gin.Context) {
	paket, ok := h.hajiPaket(c, c.Query("paket_id"))
	if !ok {
		return
	}
	list, err := h.jamaahRepo.FindByPaket(c.Request.Context(), paket.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sortByKamar(list)
	c.JSON(http.StatusOK, gin.H{"paket": paket, "jamaah": list})
}

// ---------- excel export ----------

var exportHeaders = []string{"No", "Nama", "NIK", "Jenis Kelamin", "No HP", "Rombongan", "Regu", "Ketua", "Gedung", "Kamar"}

func ketuaLabel(j model.Jamaah) string {
	var parts []string
	if j.KetuaRombongan {
		parts = append(parts, "Ketua Rombongan")
	}
	if j.KetuaRegu {
		parts = append(parts, "Ketua Regu")
	}
	return strings.Join(parts, ", ")
}

// Export godoc
// @Summary      Unduh daftar kamar manasik (Excel)
// @Description  File ini juga bisa diedit lalu di-import kembali.
// @Tags         manasik
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        paket_id  query  string  true  "ID paket haji"
// @Router       /manasik/export [get]
func (h *ManasikHandler) Export(c *gin.Context) {
	paket, ok := h.hajiPaket(c, c.Query("paket_id"))
	if !ok {
		return
	}
	list, err := h.jamaahRepo.FindByPaket(c.Request.Context(), paket.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sortByKamar(list)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Daftar Kamar"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"047857"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	textStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 49}) // "@" = text, keeps NIK intact

	for i, hdr := range exportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellStr(sheet, cell, hdr)
	}
	f.SetCellStyle(sheet, "A1", "J1", headerStyle)

	for r, j := range list {
		row := r + 2
		jk := "Laki-laki"
		if j.JenisKelamin == "perempuan" {
			jk = "Perempuan"
		}
		values := []string{"", j.Nama, j.NIK, jk, j.NoHP, j.Rombongan, j.Regu, ketuaLabel(j), j.GedungManasik, j.KamarManasik}
		f.SetCellInt(sheet, fmt.Sprintf("A%d", row), r+1)
		for col := 2; col <= len(values); col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			f.SetCellStr(sheet, cell, values[col-1])
		}
	}
	if len(list) > 0 {
		last := len(list) + 1
		f.SetCellStyle(sheet, "C2", fmt.Sprintf("C%d", last), textStyle)
		f.SetCellStyle(sheet, "E2", fmt.Sprintf("E%d", last), textStyle)
		f.SetCellStyle(sheet, "F2", fmt.Sprintf("J%d", last), textStyle)
	}
	widths := map[string]float64{"A": 5, "B": 28, "C": 20, "D": 14, "E": 16, "F": 12, "G": 10, "H": 18, "I": 14, "J": 10}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	help := "Petunjuk"
	f.NewSheet(help)
	lines := []string{
		"Cara import penempatan kamar manasik:",
		"1. Edit kolom Rombongan, Regu, Gedung, dan Kamar di sheet \"Daftar Kamar\".",
		"2. Kolom NIK wajib dan tidak boleh diubah; dipakai untuk mencocokkan jamaah.",
		"3. Sel yang dikosongkan akan MENGHAPUS nilai tersebut. Kolom yang dihapus seluruhnya tidak diubah.",
		"4. Kolom No, Nama, Jenis Kelamin, No HP, dan Ketua diabaikan saat import (ubah ketua lewat form jamaah).",
		"5. Simpan sebagai .xlsx (atau .csv), lalu import di menu Manasik > Daftar Kamar > Import Excel.",
	}
	for i, l := range lines {
		f.SetCellStr(help, fmt.Sprintf("A%d", i+1), l)
	}
	f.SetColWidth(help, "A", "A", 110)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat file excel"})
		return
	}
	filename := fmt.Sprintf("daftar-kamar-manasik-%s.xlsx", strings.ReplaceAll(strings.ToLower(paket.Label), " ", "-"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// ---------- excel / csv import ----------

type importRow struct {
	Row     int                 `json:"row"`
	NIK     string              `json:"nik"`
	Nama    string              `json:"nama"`
	Status  string              `json:"status"` // "ubah" | "sama" | "error"
	Error   string              `json:"error,omitempty"`
	Changes []model.FieldChange `json:"changes,omitempty"`
}

func normalizeHeader(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var importColumns = map[string]string{
	"nik":           "nik",
	"gedung":        "gedung",
	"gedungmanasik": "gedung",
	"kamar":         "kamar",
	"kamarmanasik":  "kamar",
	"rombongan":     "rombongan",
	"regu":          "regu",
}

// normalizeNIK handles NIKs that spreadsheets turned into numbers
// (e.g. "3.20101010101E+15" or "3201010101010001.0").
func normalizeNIK(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "'")
	if strings.ContainsAny(s, "eE.") {
		if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
			return strconv.FormatFloat(f, 'f', 0, 64)
		}
	}
	return strings.ReplaceAll(s, " ", "")
}

func readImportRows(filename string, data []byte) ([][]string, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".xlsx":
		f, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("file excel tidak bisa dibaca")
		}
		defer f.Close()
		sheet := "Daftar Kamar"
		if idx, _ := f.GetSheetIndex(sheet); idx < 0 {
			sheet = f.GetSheetName(0)
		}
		return f.GetRows(sheet, excelize.Options{RawCellValue: true})
	case ".csv":
		text := strings.TrimPrefix(string(data), "\xef\xbb\xbf")
		firstLine := text
		if i := strings.IndexAny(text, "\r\n"); i >= 0 {
			firstLine = text[:i]
		}
		r := csv.NewReader(strings.NewReader(text))
		// Excel with Indonesian locale saves CSV with ';'.
		if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
			r.Comma = ';'
		}
		r.FieldsPerRecord = -1
		return r.ReadAll()
	default:
		return nil, fmt.Errorf("format file harus .xlsx atau .csv")
	}
}

// Import godoc
// @Summary      Import penempatan manasik dari Excel/CSV
// @Description  Kolom yang dikenali: NIK (wajib), Gedung, Kamar, Rombongan, Regu. Gunakan dry_run=true untuk preview.
// @Tags         manasik
// @Accept       multipart/form-data
// @Produce      json
// @Param        paket_id  query     string  true   "ID paket haji"
// @Param        dry_run   query     bool    false  "Preview saja, tanpa menyimpan"
// @Param        file      formData  file    true   "File .xlsx atau .csv"
// @Success      200  {object}  map[string]interface{}
// @Router       /manasik/import [post]
func (h *ManasikHandler) Import(c *gin.Context) {
	paket, ok := h.hajiPaket(c, c.Query("paket_id"))
	if !ok {
		return
	}
	dryRun := c.Query("dry_run") == "true"

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImportBytes+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file wajib diunggah (maks 5MB)"})
		return
	}
	if fh.Size > maxImportBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran file maksimal 5MB"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file tidak bisa dibuka"})
		return
	}
	data, err := io.ReadAll(src)
	src.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file tidak bisa dibaca"})
		return
	}

	rows, err := readImportRows(fh.Filename, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(rows) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file kosong atau hanya berisi header"})
		return
	}
	if len(rows) > maxImportRows+1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("maksimal %d baris", maxImportRows)})
		return
	}

	// Map header columns.
	col := map[string]int{}
	for i, hdr := range rows[0] {
		if key, ok := importColumns[normalizeHeader(hdr)]; ok {
			if _, dup := col[key]; !dup {
				col[key] = i
			}
		}
	}
	if _, ok := col["nik"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kolom NIK tidak ditemukan di baris pertama"})
		return
	}
	if len(col) == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak ada kolom Gedung, Kamar, Rombongan, atau Regu"})
		return
	}
	cell := func(r []string, key string) *string {
		i, ok := col[key]
		if !ok {
			return nil // column absent: leave unchanged
		}
		v := ""
		if i < len(r) {
			v = strings.TrimSpace(r[i])
		}
		return &v
	}

	ctx := c.Request.Context()
	members, err := h.jamaahRepo.FindByPaket(ctx, paket.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	byNIK := make(map[string]*model.Jamaah, len(members))
	for i := range members {
		byNIK[members[i].NIK] = &members[i]
	}

	type pending struct {
		jamaah  *model.Jamaah
		update  repository.ManasikUpdate
		changes []model.FieldChange
	}
	var (
		results []importRow
		toApply []pending
		seen    = map[string]int{}
		nUbah   int
		nSama   int
		nError  int
	)
	for i, r := range rows[1:] {
		rowNum := i + 2
		blank := true
		for _, v := range r {
			if strings.TrimSpace(v) != "" {
				blank = false
				break
			}
		}
		if blank {
			continue
		}
		res := importRow{Row: rowNum}
		if p := cell(r, "nik"); p != nil {
			res.NIK = normalizeNIK(*p)
		}
		switch {
		case res.NIK == "":
			res.Status, res.Error = "error", "NIK kosong"
		case seen[res.NIK] != 0:
			res.Status, res.Error = "error", fmt.Sprintf("NIK duplikat (sama dengan baris %d)", seen[res.NIK])
		case byNIK[res.NIK] == nil:
			res.Status, res.Error = "error", "NIK tidak ditemukan di "+paket.Label
		}
		if res.NIK != "" && seen[res.NIK] == 0 {
			seen[res.NIK] = rowNum
		}
		if res.Status == "error" {
			nError++
			results = append(results, res)
			continue
		}
		j := byNIK[res.NIK]
		res.Nama = j.Nama
		u := repository.ManasikUpdate{
			GedungManasik: cell(r, "gedung"),
			KamarManasik:  cell(r, "kamar"),
			Rombongan:     cell(r, "rombongan"),
			Regu:          cell(r, "regu"),
		}
		res.Changes = manasikChanges(j, u)
		if len(res.Changes) == 0 {
			res.Status = "sama"
			nSama++
		} else {
			res.Status = "ubah"
			nUbah++
			toApply = append(toApply, pending{jamaah: j, update: u, changes: res.Changes})
		}
		results = append(results, res)
	}

	if !dryRun {
		for _, p := range toApply {
			if err := h.jamaahRepo.SetManasikOne(ctx, p.jamaah.ID, p.update); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gagal menyimpan %s: %v", p.jamaah.Nama, err)})
				return
			}
			logAuditWithChanges(c, h.auditRepo, "jamaah", p.jamaah.ID, p.jamaah.Nama, "update",
				fmt.Sprintf("Import penempatan manasik: %s", p.jamaah.Nama), p.changes)
		}
	}

	if results == nil {
		results = []importRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"dry_run": dryRun,
		"rows":    results,
		"summary": gin.H{"ubah": nUbah, "sama": nSama, "error": nError},
	})
}

// ---------- sesi & absensi ----------

type createSesiRequest struct {
	PaketID string `json:"paket_id" binding:"required"`
	Nama    string `json:"nama" binding:"required"`
	Tanggal string `json:"tanggal" binding:"required"` // yyyy-mm-dd
}

// ListSesi godoc
// @Summary      Daftar sesi manasik
// @Tags         manasik
// @Produce      json
// @Param        paket_id  query  string  true  "ID paket haji"
// @Success      200  {array}  model.SesiManasik
// @Router       /manasik/sesi [get]
func (h *ManasikHandler) ListSesi(c *gin.Context) {
	paket, ok := h.hajiPaket(c, c.Query("paket_id"))
	if !ok {
		return
	}
	list, err := h.repo.FindSesiByPaket(c.Request.Context(), paket.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateSesi godoc
// @Summary      Tambah sesi manasik
// @Tags         manasik
// @Accept       json
// @Produce      json
// @Param        body  body  createSesiRequest  true  "Sesi"
// @Success      201  {object}  model.SesiManasik
// @Router       /manasik/sesi [post]
func (h *ManasikHandler) CreateSesi(c *gin.Context) {
	var req createSesiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paket, nama, dan tanggal sesi wajib diisi"})
		return
	}
	paket, ok := h.hajiPaket(c, req.PaketID)
	if !ok {
		return
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" || len(nama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama sesi wajib diisi (maks 100 karakter)"})
		return
	}
	tanggal, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format tanggal tidak valid"})
		return
	}
	s := &model.SesiManasik{PaketID: paket.ID, Nama: nama, Tanggal: tanggal}
	if err := h.repo.CreateSesi(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logAudit(c, h.auditRepo, "manasik", s.ID, s.Nama, "create",
		fmt.Sprintf("Menambah sesi manasik %s: %s", paket.Label, s.Nama))
	c.JSON(http.StatusCreated, s)
}

func (h *ManasikHandler) loadSesi(c *gin.Context) (*model.SesiManasik, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id sesi tidak valid"})
		return nil, false
	}
	s, err := h.repo.FindSesiByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "sesi tidak ditemukan"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return nil, false
	}
	return s, true
}

// DeleteSesi godoc
// @Summary      Hapus sesi manasik beserta data kehadirannya
// @Tags         manasik
// @Param        id  path  string  true  "ID sesi"
// @Success      200  {object}  map[string]string
// @Router       /manasik/sesi/{id} [delete]
func (h *ManasikHandler) DeleteSesi(c *gin.Context) {
	s, ok := h.loadSesi(c)
	if !ok {
		return
	}
	if err := h.repo.DeleteSesi(c.Request.Context(), s.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logAudit(c, h.auditRepo, "manasik", s.ID, s.Nama, "delete", fmt.Sprintf("Menghapus sesi manasik: %s", s.Nama))
	c.JSON(http.StatusOK, gin.H{"message": "sesi berhasil dihapus"})
}

type kehadiranJamaah struct {
	ID            primitive.ObjectID      `json:"id"`
	Nama          string                  `json:"nama"`
	NIK           string                  `json:"nik"`
	JenisKelamin  string                  `json:"jenis_kelamin"`
	GedungManasik string                  `json:"gedung_manasik"`
	KamarManasik  string                  `json:"kamar_manasik"`
	Rombongan     string                  `json:"rombongan"`
	Regu          string                  `json:"regu"`
	Hadir         *model.KehadiranManasik `json:"hadir"`
}

func toKehadiranJamaah(j *model.Jamaah, k *model.KehadiranManasik) kehadiranJamaah {
	return kehadiranJamaah{
		ID: j.ID, Nama: j.Nama, NIK: j.NIK, JenisKelamin: j.JenisKelamin,
		GedungManasik: j.GedungManasik, KamarManasik: j.KamarManasik,
		Rombongan: j.Rombongan, Regu: j.Regu, Hadir: k,
	}
}

// GetSesi godoc
// @Summary      Detail sesi manasik dengan status kehadiran tiap jamaah
// @Tags         manasik
// @Produce      json
// @Param        id  path  string  true  "ID sesi"
// @Success      200  {object}  map[string]interface{}
// @Router       /manasik/sesi/{id} [get]
func (h *ManasikHandler) GetSesi(c *gin.Context) {
	s, ok := h.loadSesi(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	paket, err := h.paketRepo.FindByID(ctx, s.PaketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "paket sesi tidak ditemukan"})
		return
	}
	members, err := h.jamaahRepo.FindByPaket(ctx, s.PaketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	hadir, err := h.repo.FindKehadiran(ctx, s.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	byJamaah := make(map[primitive.ObjectID]*model.KehadiranManasik, len(hadir))
	for i := range hadir {
		byJamaah[hadir[i].JamaahID] = &hadir[i]
	}
	out := make([]kehadiranJamaah, 0, len(members))
	for i := range members {
		out = append(out, toKehadiranJamaah(&members[i], byJamaah[members[i].ID]))
	}
	s.JumlahHadir = int64(len(hadir))
	c.JSON(http.StatusOK, gin.H{"sesi": s, "paket": paket, "jamaah": out})
}

type markHadirRequest struct {
	Token    string `json:"token"`
	JamaahID string `json:"jamaah_id"`
}

// MarkHadir godoc
// @Summary      Catat kehadiran (scan QR atau manual)
// @Tags         manasik
// @Accept       json
// @Produce      json
// @Param        id    path  string            true  "ID sesi"
// @Param        body  body  markHadirRequest  true  "token dari QR, atau jamaah_id untuk manual"
// @Success      200  {object}  map[string]interface{}
// @Router       /manasik/sesi/{id}/hadir [post]
func (h *ManasikHandler) MarkHadir(c *gin.Context) {
	s, ok := h.loadSesi(c)
	if !ok {
		return
	}
	var req markHadirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid"})
		return
	}

	var (
		jamaahID primitive.ObjectID
		metode   string
	)
	switch {
	case req.Token != "":
		id, valid := parseManasikToken(h.qrSecret, req.Token)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "QR tidak valid. Minta jamaah membuka ulang QR dari halaman Cek Kamar."})
			return
		}
		jamaahID, metode = id, "qr"
	case req.JamaahID != "":
		id, err := primitive.ObjectIDFromHex(req.JamaahID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id jamaah tidak valid"})
			return
		}
		jamaahID, metode = id, "manual"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "token atau jamaah_id wajib diisi"})
		return
	}

	ctx := c.Request.Context()
	j, err := h.jamaahRepo.FindByID(ctx, jamaahID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "jamaah tidak ditemukan"})
		return
	}
	if j.PaketID != s.PaketID {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s bukan peserta paket sesi ini", j.Nama)})
		return
	}

	k, created, err := h.repo.MarkHadir(ctx, &model.KehadiranManasik{
		SesiID: s.ID, JamaahID: j.ID, DicatatOleh: adminUsername(c), Metode: metode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	status := "baru"
	if !created {
		status = "sudah"
	}
	c.JSON(http.StatusOK, gin.H{"status": status, "jamaah": toKehadiranJamaah(j, k)})
}

// UnmarkHadir godoc
// @Summary      Batalkan kehadiran jamaah di sesi
// @Tags         manasik
// @Param        id         path  string  true  "ID sesi"
// @Param        jamaahId   path  string  true  "ID jamaah"
// @Success      200  {object}  map[string]string
// @Router       /manasik/sesi/{id}/hadir/{jamaahId} [delete]
func (h *ManasikHandler) UnmarkHadir(c *gin.Context) {
	s, ok := h.loadSesi(c)
	if !ok {
		return
	}
	jamaahID, err := primitive.ObjectIDFromHex(c.Param("jamaahId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id jamaah tidak valid"})
		return
	}
	deleted, err := h.repo.DeleteKehadiran(c.Request.Context(), s.ID, jamaahID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "data kehadiran tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kehadiran dibatalkan"})
}

func (h *ManasikHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/manasik")
	{
		g.GET("/jamaah", h.ListJamaah)
		g.GET("/export", h.Export)
		g.POST("/import", h.Import)
		g.GET("/sesi", h.ListSesi)
		g.POST("/sesi", h.CreateSesi)
		g.GET("/sesi/:id", h.GetSesi)
		g.DELETE("/sesi/:id", h.DeleteSesi)
		g.POST("/sesi/:id/hadir", h.MarkHadir)
		g.DELETE("/sesi/:id/hadir/:jamaahId", h.UnmarkHadir)
	}
}
