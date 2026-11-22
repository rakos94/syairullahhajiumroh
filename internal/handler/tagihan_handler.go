package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"syairullahhajiumroh/internal/model"
	"syairullahhajiumroh/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TagihanHandler struct {
	jamaahRepo *repository.JamaahRepository
	paketRepo  *repository.PaketRepository
}

func NewTagihanHandler(jamaahRepo *repository.JamaahRepository, paketRepo *repository.PaketRepository) *TagihanHandler {
	return &TagihanHandler{jamaahRepo: jamaahRepo, paketRepo: paketRepo}
}

type tagihanRow struct {
	ID               primitive.ObjectID `json:"id"`
	Nama             string             `json:"nama"`
	NIK              string             `json:"nik"`
	NoHP             string             `json:"no_hp"`
	Paket            string             `json:"paket"`
	StatusPembayaran string             `json:"status_pembayaran"`
	model.Tagihan
	// Peringatan flags a mismatch between the manual payment status and the receipts.
	Peringatan string `json:"peringatan,omitempty"`
}

type tagihanSummary struct {
	JumlahJamaah     int     `json:"jumlah_jamaah"`
	TotalTagihan     float64 `json:"total_tagihan"`
	TotalDibayar     float64 `json:"total_dibayar"`
	TotalSisa        float64 `json:"total_sisa"`
	JumlahBelumLunas int     `json:"jumlah_belum_lunas"`
	HargaBelumDiatur int     `json:"harga_belum_diatur"`
}

func peringatanTagihan(status string, t *model.Tagihan) string {
	if !t.HargaDiatur {
		return "Harga paket belum diatur"
	}
	switch {
	case status == "lunas" && t.Sisa > 0:
		return "Status Lunas tetapi masih ada sisa tagihan"
	case status != "lunas" && t.Sisa <= 0:
		return "Pembayaran sudah penuh tetapi status belum Lunas"
	case t.Sisa < 0:
		return "Kelebihan bayar"
	}
	return ""
}

// build loads jamaah (optionally of one paket), computes bills and filters.
func (h *TagihanHandler) build(c *gin.Context) ([]tagihanRow, tagihanSummary, bool) {
	ctx := c.Request.Context()
	var paketID *primitive.ObjectID
	if s := c.Query("paket_id"); s != "" {
		id, err := primitive.ObjectIDFromHex(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "paket_id tidak valid"})
			return nil, tagihanSummary{}, false
		}
		paketID = &id
	}
	hanyaBelumLunas := c.Query("belum_lunas") == "true"

	list, err := h.jamaahRepo.FindActive(ctx, paketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, tagihanSummary{}, false
	}
	pakets, err := h.paketRepo.FindAll(ctx, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, tagihanSummary{}, false
	}
	paketMap := make(map[primitive.ObjectID]*model.Paket, len(pakets))
	for i := range pakets {
		paketMap[pakets[i].ID] = &pakets[i]
	}

	rows := []tagihanRow{}
	var sum tagihanSummary
	for i := range list {
		j := &list[i]
		p := paketMap[j.PaketID]
		t := model.HitungTagihan(j, p)
		belumLunas := !t.HargaDiatur || t.Sisa > 0
		if hanyaBelumLunas && !belumLunas {
			continue
		}
		label := ""
		if p != nil {
			label = p.Label
		}
		rows = append(rows, tagihanRow{
			ID: j.ID, Nama: j.Nama, NIK: j.NIK, NoHP: j.NoHP, Paket: label,
			StatusPembayaran: j.StatusPembayaran, Tagihan: *t,
			Peringatan: peringatanTagihan(j.StatusPembayaran, t),
		})
		sum.JumlahJamaah++
		sum.TotalDibayar += t.TotalDibayar
		if t.HargaDiatur {
			sum.TotalTagihan += t.Harga
			if t.Sisa > 0 {
				sum.TotalSisa += t.Sisa
				sum.JumlahBelumLunas++
			}
		} else {
			sum.HargaBelumDiatur++
		}
	}
	// Largest outstanding balance first; unpriced last.
	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].HargaDiatur != rows[b].HargaDiatur {
			return rows[a].HargaDiatur
		}
		return rows[a].Sisa > rows[b].Sisa
	})
	return rows, sum, true
}

// List godoc
// @Summary      Laporan tagihan jamaah
// @Description  Harga paket dikurangi total bukti DP dan pelunasan.
// @Tags         tagihan
// @Produce      json
// @Param        paket_id     query  string  false  "Filter paket"
// @Param        belum_lunas  query  bool    false  "Hanya yang masih punya sisa"
// @Success      200  {object}  map[string]interface{}
// @Router       /tagihan [get]
func (h *TagihanHandler) List(c *gin.Context) {
	rows, sum, ok := h.build(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": sum, "data": rows})
}

var statusPembayaranLabel = map[string]string{"belum_bayar": "Belum Bayar", "dp": "DP", "lunas": "Lunas"}

// Export godoc
// @Summary      Unduh laporan tagihan (Excel)
// @Tags         tagihan
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        paket_id     query  string  false  "Filter paket"
// @Param        belum_lunas  query  bool    false  "Hanya yang masih punya sisa"
// @Router       /tagihan/export [get]
func (h *TagihanHandler) Export(c *gin.Context) {
	rows, sum, ok := h.build(c)
	if !ok {
		return
	}
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Tagihan"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"047857"}, Pattern: 1},
	})
	rupiah := "\"Rp\" #,##0;[Red]-\"Rp\" #,##0"
	moneyStyle, _ := f.NewStyle(&excelize.Style{CustomNumFmt: &rupiah})
	textStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 49})
	boldMoney, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, CustomNumFmt: &rupiah})

	// Columns: A No, B Nama, C NIK, D No HP, E Paket, F Pilihan Harga, G Status,
	// H Harga dasar, I Potongan, J Harga, K Dibayar, L Sisa, M Catatan
	headers := []string{"No", "Nama", "NIK", "No HP", "Paket", "Pilihan Harga", "Status", "Harga Dasar", "Potongan", "Harga", "Dibayar", "Sisa", "Catatan"}
	for i, hdr := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellStr(sheet, cell, hdr)
	}
	f.SetCellStyle(sheet, "A1", "M1", headerStyle)

	var totalPotongan float64
	for i, r := range rows {
		row := i + 2
		cell := func(col string) string { return fmt.Sprintf("%s%d", col, row) }
		f.SetCellInt(sheet, cell("A"), i+1)
		f.SetCellStr(sheet, cell("B"), r.Nama)
		f.SetCellStr(sheet, cell("C"), r.NIK)
		f.SetCellStr(sheet, cell("D"), r.NoHP)
		f.SetCellStr(sheet, cell("E"), r.Paket)
		f.SetCellStr(sheet, cell("F"), r.PilihanHarga)
		f.SetCellStr(sheet, cell("G"), statusPembayaranLabel[r.StatusPembayaran])
		if r.HargaDiatur {
			f.SetCellFloat(sheet, cell("H"), r.HargaDasar, 0, 64)
			if r.Potongan > 0 {
				f.SetCellFloat(sheet, cell("I"), r.Potongan, 0, 64)
				totalPotongan += r.Potongan
			}
			f.SetCellFloat(sheet, cell("J"), r.Harga, 0, 64)
			f.SetCellFloat(sheet, cell("L"), r.Sisa, 0, 64)
		}
		f.SetCellFloat(sheet, cell("K"), r.TotalDibayar, 0, 64)
		f.SetCellStr(sheet, cell("M"), r.Peringatan)
	}
	last := len(rows) + 1
	if len(rows) > 0 {
		f.SetCellStyle(sheet, "C2", fmt.Sprintf("D%d", last), textStyle)
		f.SetCellStyle(sheet, "H2", fmt.Sprintf("L%d", last), moneyStyle)
	}
	total := last + 2
	f.SetCellStr(sheet, fmt.Sprintf("G%d", total), "Total")
	f.SetCellFloat(sheet, fmt.Sprintf("I%d", total), totalPotongan, 0, 64)
	f.SetCellFloat(sheet, fmt.Sprintf("J%d", total), sum.TotalTagihan, 0, 64)
	f.SetCellFloat(sheet, fmt.Sprintf("K%d", total), sum.TotalDibayar, 0, 64)
	f.SetCellFloat(sheet, fmt.Sprintf("L%d", total), sum.TotalSisa, 0, 64)
	f.SetCellStyle(sheet, fmt.Sprintf("G%d", total), fmt.Sprintf("L%d", total), boldMoney)

	for col, w := range map[string]float64{"A": 5, "B": 26, "C": 20, "D": 16, "E": 18, "F": 14, "G": 12, "H": 16, "I": 14, "J": 16, "K": 16, "L": 16, "M": 44} {
		f.SetColWidth(sheet, col, col, w)
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat file excel"})
		return
	}
	name := "laporan-tagihan"
	if c.Query("belum_lunas") == "true" {
		name += "-belum-lunas"
	}
	name = fmt.Sprintf("%s-%s.xlsx", name, time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ToLower(name)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func (h *TagihanHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/tagihan", h.List)
	r.GET("/tagihan/export", h.Export)
}
