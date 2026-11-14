package model

// Tagihan is the computed bill of a jamaah: price (paket price or the chosen
// price option, minus any discount) vs. total of the uploaded DP and
// pelunasan receipts.
type Tagihan struct {
	// HargaDasar is the paket price or the chosen price option, before discount.
	HargaDasar   float64 `json:"harga_dasar"`
	PilihanHarga string  `json:"pilihan_harga,omitempty"`
	Potongan     float64 `json:"potongan"`
	// Harga is what this jamaah has to pay: HargaDasar - Potongan.
	Harga        float64 `json:"harga"`
	TotalDibayar float64 `json:"total_dibayar"`
	Sisa         float64 `json:"sisa"`
	// HargaDiatur is false when no price is set yet, so Sisa is meaningless.
	HargaDiatur bool `json:"harga_diatur"`
}

// HargaDasarJamaah returns the price before discount for j in paket p and
// the name of the price option used (empty when the paket price is used).
func HargaDasarJamaah(j *Jamaah, p *Paket) (float64, string) {
	if p == nil {
		return 0, ""
	}
	if j.PilihanHarga != "" {
		if opt, ok := p.FindPilihanHarga(j.PilihanHarga); ok {
			return opt.Harga, opt.Nama
		}
	}
	return p.Harga, ""
}

// HitungTagihan computes the bill of j for paket p (p may be nil).
func HitungTagihan(j *Jamaah, p *Paket) *Tagihan {
	t := &Tagihan{}
	for _, b := range j.BuktiDP {
		t.TotalDibayar += b.Nominal
	}
	for _, b := range j.BuktiPelunasan {
		t.TotalDibayar += b.Nominal
	}
	t.HargaDasar, t.PilihanHarga = HargaDasarJamaah(j, p)
	if t.HargaDasar > 0 {
		t.HargaDiatur = true
		t.Potongan = j.Potongan
		t.Harga = t.HargaDasar - t.Potongan
		if t.Harga < 0 {
			t.Harga = 0
		}
		t.Sisa = t.Harga - t.TotalDibayar
	}
	return t
}
