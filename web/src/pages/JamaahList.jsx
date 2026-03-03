import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { fetchJamaah, deleteJamaah, fetchPaketList } from '../api';
import StatusBadge from '../components/StatusBadge';

const PAGE_SIZE = 10;

export default function JamaahList() {
  const [jamaahList, setJamaahList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filterPaket, setFilterPaket] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterKelengkapan, setFilterKelengkapan] = useState('');
  const [paketOptions, setPaketOptions] = useState([]);
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();

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
        </select>
      </div>

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
                    Status
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
                    <td className="px-6 py-4 whitespace-nowrap">
                      <StatusBadge status={j.status_pembayaran} />
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
