const rupiahFmt = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 });

// 25000000 -> "Rp 25.000.000"
export function formatRupiah(value) {
  return rupiahFmt.format(Number(value) || 0).replace(/\u00a0/g, ' ');
}

const SATUAN = ['', 'satu', 'dua', 'tiga', 'empat', 'lima', 'enam', 'tujuh', 'delapan', 'sembilan', 'sepuluh', 'sebelas'];

function terbilangRaw(n) {
  if (n < 12) return SATUAN[n];
  if (n < 20) return `${terbilangRaw(n - 10)} belas`;
  const part = (div, word) => {
    const head = Math.floor(n / div);
    const rest = n % div;
    return `${terbilangRaw(head)} ${word}${rest ? ` ${terbilangRaw(rest)}` : ''}`;
  };
  if (n < 100) return part(10, 'puluh');
  if (n < 200) return `seratus${n - 100 ? ` ${terbilangRaw(n - 100)}` : ''}`;
  if (n < 1000) return part(100, 'ratus');
  if (n < 2000) return `seribu${n - 1000 ? ` ${terbilangRaw(n - 1000)}` : ''}`;
  if (n < 1e6) return part(1e3, 'ribu');
  if (n < 1e9) return part(1e6, 'juta');
  if (n < 1e12) return part(1e9, 'miliar');
  return part(1e12, 'triliun');
}

// 3000000 -> "tiga juta rupiah" (helps catch a missing/extra zero).
export function terbilang(value) {
  const n = Math.floor(Number(value) || 0);
  if (n <= 0) return '';
  return `${terbilangRaw(n).replace(/\s+/g, ' ').trim()} rupiah`;
}
