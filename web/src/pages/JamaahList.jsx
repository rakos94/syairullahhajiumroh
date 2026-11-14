import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { fetchJamaah, deleteJamaah, fetchPaketList, bulkSetManasik } from '../api';
import StatusBadge from '../components/StatusBadge';
import { formatRupiah } from '../utils/format';

const PAGE_SIZE = 10;

export default function JamaahList() {
  // Filters can be preset via URL, e.g. /jamaah?paket_id=...&kelengkapan=manasik_belum
  const [searchParams] = useSearchParams();
  const [jamaahList, setJamaahList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filterPaket, setFilterPaket] = useState(() => searchParams.get('paket_id') || '');
  const [filterStatus, setFilterStatus] = useState(() => searchParams.get('status_pembayaran') || '');
  const [filterKelengkapan, setFilterKelengkapan] = useState(() => searchParams.get('kelengkapan') || '');
  const [paketOptions, setPaketOptions] = useState([]);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();

  // Bulk manasik assignment. Selection is keyed by id and survives paging/filtering.
  const [selected, setSelected] = useState({});
  const [bulkGedung, setBulkGedung] = useState('');
  const [bulkKamar, setBulkKamar] = useState('');
  const [bulkRombongan, setBulkRombongan] = useState('');
  const [bulkRegu, setBulkRegu] = useState('');
  const [bulkSaving, setBulkSaving] = useState(false);
  const [bulkMsg, setBulkMsg] = useState('');
  const selectedIds = Object.keys(selected);

  useEffect(() => {
    fetchPaketList().then(setPaketOptions).catch(() => {});
  }, []);

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  const loadData = async () => {
    try {
      setLoading(true);
      const res = await fetchJamaah({ paket_id: filterPaket, search: debouncedSearch, status_pembayaran: filterStatus, kelengkapan: filterKelengkapan, page, limit: PAGE_SIZE });
      setJamaahList(res.data || []);
      setTotalPages(res.total_pages || 1);
      setTotal(res.total || 0);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [filterPaket, filterStatus, filterKelengkapan, debouncedSearch, page]);

  // Reset to page 1 when filter or search changes
  useEffect(() => {
    setPage(1);
  }, [filterPaket, filterStatus, filterKelengkapan, debouncedSearch]);

  const handleDelete = async (id, nama) => {
    if (!confirm(`Hapus jamaah "${nama}"?`)) return;
    try {
      await deleteJamaah(id);
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const isHaji = (j) => j.paket?.tipe === 'haji';
  const hajiOnPage = jamaahList.filter(isHaji);
  const allPageSelected = hajiOnPage.length > 0 && hajiOnPage.every((j) => selected[j.id]);

  const toggleOne = (j) => {
    setSelected((prev) => {
      const next = { ...prev };
      if (next[j.id]) delete next[j.id];
      else next[j.id] = j.nama;
      return next;
    });
  };

  const togglePage = () => {
    setSelected((prev) => {
      const next = { ...prev };
      hajiOnPage.forEach((j) => {
        if (allPageSelected) delete next[j.id];
        else next[j.id] = j.nama;
      });
      return next;
    });
  };

  const applyBulk = async (clear = false) => {
    const payload = clear
      ? { gedung_manasik: '', kamar_manasik: '', rombongan: '', regu: '' }
      : {
          gedung_manasik: bulkGedung.trim() || undefined,
          kamar_manasik: bulkKamar.trim() || undefined,
          rombongan: bulkRombongan.trim() || undefined,
          regu: bulkRegu.trim() || undefined,
        };
    if (!clear && Object.values(payload).every((v) => v === undefined)) {
      setBulkMsg('Isi minimal satu: gedung, kamar, rombongan, atau regu');
      return;
    }
    if (clear && !confirm(`Hapus gedung, kamar, rombongan & regu untuk ${selectedIds.length} jamaah?`)) return;

    setBulkSaving(true);
    setBulkMsg('');
    try {
      const res = await bulkSetManasik(selectedIds, payload);
      setSelected({});
      setBulkGedung('');
      setBulkKamar('');
      setBulkRombongan('');
      setBulkRegu('');
      setBulkMsg(res.message);
      loadData();
    } catch (err) {
      setBulkMsg(err.message);
    } finally {
      setBulkSaving(false);
    }
  };

  const bulkInputClass =
    'px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 w-full sm:w-28';

  return (
    <div>
      <div className="sm:flex sm:items-center sm:justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Daftar Jamaah</h1>
        <Link
          to="/jamaah/new"
          className="mt-3 sm:mt-0 inline-flex items-center px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700"
        >
          + Tambah Jamaah
        </Link>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 mb-4">
        <input
          type="text"
          placeholder="Cari nama atau NIK..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
        />
        <select
          value={filterPaket}
          onChange={(e) => setFilterPaket(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
        >
          <option value="">Semua Paket</option>
          {paketOptions.map((p) => (
            <option key={p.id} value={p.id}>
              {p.label}
            </option>
          ))}
        </select>
        <select
          value={filterStatus}
          onChange={(e) => setFilterStatus(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
        >
          <option value="">Semua Status</option>
          <option value="belum_bayar">Belum Bayar</option>
          <option value="dp">DP</option>
          <option value="lunas">Lunas</option>
        </select>
        <select
          value={filterKelengkapan}
          onChange={(e) => setFilterKelengkapan(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
        >
          <option value="">Semua Kelengkapan</option>
          <option value="batik_nasional_belum">Batik Nasional Belum</option>
          <option value="batik_kbih_belum">Batik KBIH Belum</option>
          <option value="koper_belum">Koper Belum</option>
          <option value="manasik_belum">Kamar Manasik Belum</option>
        </select>
      </div>

      {selectedIds.length > 0 && (
        <div className="mb-4 p-4 bg-emerald-50 border border-emerald-200 rounded-lg">
          <div className="flex flex-col sm:flex-row sm:flex-wrap sm:items-center gap-3">
            <p className="text-sm font-medium text-emerald-900 whitespace-nowrap">
              {selectedIds.length} jamaah dipilih
            </p>
            <input
              value={bulkGedung}
              onChange={(e) => setBulkGedung(e.target.value)}
              placeholder="Gedung"
              aria-label="Gedung manasik"
              className={bulkInputClass}
            />
            <input
              value={bulkKamar}
              onChange={(e) => setBulkKamar(e.target.value)}
              placeholder="Kamar"
              aria-label="Kamar manasik"
              className={bulkInputClass}
            />
            <input
              value={bulkRombongan}
              onChange={(e) => setBulkRombongan(e.target.value)}
              placeholder="Rombongan"
              aria-label="Rombongan"
              className={bulkInputClass}
            />
            <input
              value={bulkRegu}
              onChange={(e) => setBulkRegu(e.target.value)}
              placeholder="Regu"
              aria-label="Regu"
              className={bulkInputClass}
            />
            <div className="flex flex-wrap gap-2">
              <button
                onClick={() => applyBulk(false)}
                disabled={bulkSaving}
                className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
              >
                {bulkSaving ? 'Menyimpan...' : 'Terapkan'}
              </button>
              <button
                onClick={() => applyBulk(true)}
                disabled={bulkSaving}
                className="px-4 py-2 bg-white text-red-600 text-sm font-medium rounded-md border border-red-200 hover:bg-red-50 disabled:opacity-50"
              >
                Hapus Penempatan
              </button>
              <button
                onClick={() => setSelected({})}
                disabled={bulkSaving}
                className="px-4 py-2 text-gray-600 text-sm font-medium rounded-md hover:bg-gray-100"
              >
                Batal
              </button>
            </div>
          </div>
          <p className="mt-2 text-xs text-emerald-800">
            Kolom yang dikosongkan tidak diubah. Pilihan tetap tersimpan saat pindah halaman atau
            mengganti filter.
          </p>
        </div>
      )}

      {bulkMsg && (
        <div className="mb-4 p-3 bg-blue-50 text-blue-800 rounded-md text-sm flex justify-between gap-3">
          <span>{bulkMsg}</span>
          <button onClick={() => setBulkMsg('')} className="text-blue-600 hover:text-blue-900" aria-label="Tutup">
            ✕
          </button>
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <p className="text-gray-500">Memuat data...</p>
      ) : jamaahList.length === 0 ? (
        <p className="text-gray-500">Tidak ada data jamaah.</p>
      ) : (
        <>
          <div className="overflow-x-auto bg-white rounded-lg shadow">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="pl-4 py-3 w-8">
                    <input
                      type="checkbox"
                      checked={allPageSelected}
                      disabled={hajiOnPage.length === 0}
                      onChange={togglePage}
                      title="Pilih semua jamaah haji di halaman ini"
                      aria-label="Pilih semua jamaah haji di halaman ini"
                      className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500 disabled:opacity-30"
                    />
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase w-12">
                    No
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Nama
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    NIK
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Paket
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Gedung / Kamar
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                    Sisa Tagihan
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                    Aksi
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {jamaahList.map((j, idx) => (
                  <tr
                    key={j.id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/jamaah/${j.id}`)}
                  >
                    <td className="pl-4 py-4" onClick={(e) => e.stopPropagation()}>
                      {isHaji(j) && (
                        <input
                          type="checkbox"
                          checked={Boolean(selected[j.id])}
                          onChange={() => toggleOne(j)}
                          aria-label={`Pilih ${j.nama}`}
                          className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                        />
                      )}
                    </td>
                    <td className="px-4 py-4 whitespace-nowrap text-sm text-gray-400">
                      {(page - 1) * PAGE_SIZE + idx + 1}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {j.nama}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {j.nik}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {j.paket?.label || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      {!isHaji(j) ? (
                        <span className="text-gray-300">-</span>
                      ) : j.gedung_manasik || j.kamar_manasik ? (
                        <span className="text-gray-700">
                          {j.gedung_manasik || '?'} / {j.kamar_manasik || '?'}
                        </span>
                      ) : (
                        <span className="text-amber-600">Belum</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={j.status_pembayaran} />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                      {!j.tagihan?.harga_diatur ? (
                        <span className="text-gray-300">-</span>
                      ) : j.tagihan.sisa > 0 ? (
                        <span className="text-gray-900">{formatRupiah(j.tagihan.sisa)}</span>
                      ) : (
                        <span className="text-emerald-700">Lunas</span>
                      )}
                    </td>
                    <td
                      className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-2"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <button
                        onClick={() => navigate(`/jamaah/${j.id}/edit`)}
                        className="text-blue-600 hover:text-blue-800 font-medium"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(j.id, j.nama)}
                        className="text-red-600 hover:text-red-800 font-medium"
                      >
                        Hapus
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-gray-500">
                Menampilkan {(page - 1) * PAGE_SIZE + 1}-{Math.min(page * PAGE_SIZE, total)} dari {total} jamaah
              </p>
              <div className="flex gap-1">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  Prev
                </button>
                {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={`px-3 py-1.5 text-sm border rounded-md ${
                      p === page
                        ? 'bg-emerald-600 text-white border-emerald-600'
                        : 'border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    {p}
                  </button>
                ))}
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
