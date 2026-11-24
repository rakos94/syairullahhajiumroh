// Items a jamaah receives, per paket tipe: haji gets batik, umroh a seragam;
// both get a koper. Unknown tipe (no paket chosen yet) shows koper only.
const KOPER = { field: 'koper_sudah_diterima', label: 'Koper Sudah Diterima' };

export function kelengkapanItems(tipe) {
  if (tipe === 'haji') {
    return [
      { field: 'batik_nasional_sudah_dijahit', label: 'Batik Nasional Sudah Dijahit' },
      { field: 'batik_kbih_sudah_diterima', label: 'Batik KBIH Sudah Diterima' },
      KOPER,
    ];
  }
  if (tipe === 'umroh') {
    return [{ field: 'seragam_umroh_sudah_diterima', label: 'Seragam Umroh Sudah Diterima' }, KOPER];
  }
  return [KOPER];
}
