const rupiahFmt = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 });

// 25000000 -> "Rp 25.000.000"
export function formatRupiah(value) {
  return rupiahFmt.format(Number(value) || 0).replace(/\u00a0/g, ' ');
}
