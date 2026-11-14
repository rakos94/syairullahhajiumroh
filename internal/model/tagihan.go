package model

// Tagihan is the computed bill of a jamaah: paket price vs. total of the
// uploaded DP and pelunasan receipts.
type Tagihan struct {
	Harga        float64 `json:"harga"`
	TotalDibayar float64 `json:"total_dibayar"`
	Sisa         float64 `json:"sisa"`
	// HargaDiatur is false when the paket has no price yet, so Sisa is meaningless.
	HargaDiatur bool `json:"harga_diatur"`
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
	if p != nil && p.Harga > 0 {
		t.Harga = p.Harga
		t.HargaDiatur = true
		t.Sisa = p.Harga - t.TotalDibayar
	}
	return t
}
