import { useEffect, useState } from 'react';
import { fetchPaketList, createPaket, updatePaket, deletePaket } from '../api';

const bulanOptions = [
  { value: 1, label: 'Januari' },
  { value: 2, label: 'Februari' },
  { value: 3, label: 'Maret' },
  { value: 4, label: 'April' },
  { value: 5, label: 'Mei' },
  { value: 6, label: 'Juni' },
  { value: 7, label: 'Juli' },
  { value: 8, label: 'Agustus' },
  { value: 9, label: 'September' },
  { value: 10, label: 'Oktober' },
  { value: 11, label: 'November' },
  { value: 12, label: 'Desember' },
];

const emptyForm = { tipe: 'haji', tahun: new Date().getFullYear(), bulan: 0, tanggal_keberangkatan: [] };

export default function PaketList() {
  const [paketList, setPaketList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);

  const loadData = async () => {
    try {
      setLoading(true);
      const data = await fetchPaketList();
      setPaketList(data || []);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const resetForm = () => {
    setForm(emptyForm);
    setEditingId(null);
    setShowForm(false);
  };

  const handleEdit = (paket) => {
    setForm({
      tipe: paket.tipe,
      tahun: paket.tahun,
      bulan: paket.bulan || 0,
      tanggal_keberangkatan: (paket.tanggal_keberangkatan || []).map((tk) => ({
        nama: tk.nama,
        tanggal: tk.tanggal ? tk.tanggal.slice(0, 10) : '',
      })),
    });
    setEditingId(paket.id);
    setShowForm(true);
  };

  const handleDelete = async (id, label) => {
    if (!confirm(`Hapus paket "${label}"?`)) return;
    try {
      await deletePaket(id);
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSaving(true);

    const payload = {
      tipe: form.tipe,
      tahun: Number(form.tahun),
      bulan: form.tipe === 'umroh' ? Number(form.bulan) : 0,
      tanggal_keberangkatan:
        form.tanggal_keberangkatan.length > 0
          ? form.tanggal_keberangkatan
              .filter((tk) => tk.nama)
              .map((tk) => ({
                nama: tk.nama,
                ...(tk.tanggal ? { tanggal: new Date(tk.tanggal + 'T00:00:00Z').toISOString() } : {}),
              }))
          : undefined,
    };

    try {
      if (editingId) {
        await updatePaket(editingId, payload);
      } else {
        await createPaket(payload);
      }
      resetForm();
      loadData();
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const addKeberangkatan = () => {
    setForm((prev) => ({
      ...prev,
      tanggal_keberangkatan: [...prev.tanggal_keberangkatan, { nama: '', tanggal: '' }],
    }));
  };

  const updateKeberangkatan = (index, field, value) => {
    setForm((prev) => ({
      ...prev,
      tanggal_keberangkatan: prev.tanggal_keberangkatan.map((tk, i) =>
        i === index ? { ...tk, [field]: value } : tk
      ),
    }));
  };

  const removeKeberangkatan = (index) => {
    setForm((prev) => ({
      ...prev,
      tanggal_keberangkatan: prev.tanggal_keberangkatan.filter((_, i) => i !== index),
    }));
  };

  const formatDisplayDate = (isoStr) => {
    if (!isoStr) return '-';
    const d = new Date(isoStr);
    if (isNaN(d.getTime())) return '-';
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: '2-digit', year: 'numeric' });
  };

  const inputClass =
    'block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500';

  return (
    <div>
      <div className="sm:flex sm:items-center sm:justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Daftar Paket</h1>
        {!showForm && (
          <button
            onClick={() => { resetForm(); setShowForm(true); }}
            className="mt-3 sm:mt-0 inline-flex items-center px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700"
          >
            + Tambah Paket
          </button>
        )}
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      {showForm && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            {editingId ? 'Edit Paket' : 'Tambah Paket Baru'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="flex flex-col sm:flex-row gap-3 items-end">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tipe</label>
                <select
                  value={form.tipe}
                  onChange={(e) => setForm({
                    ...form,
                    tipe: e.target.value,
                    bulan: e.target.value === 'haji' ? 0 : form.bulan || 1,
                  })}
                  className={inputClass}
                >
                  <option value="haji">Haji</option>
                  <option value="umroh">Umroh</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Tahun</label>
                <input
                  type="number"
                  value={form.tahun}
                  onChange={(e) => setForm({ ...form, tahun: e.target.value })}
                  min={2020}
                  max={2100}
                  required
                  className={inputClass}
                />
              </div>
              {form.tipe === 'umroh' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Bulan</label>
                  <select
                    value={form.bulan}
                    onChange={(e) => setForm({ ...form, bulan: Number(e.target.value) })}
                    className={inputClass}
                  >
                    {bulanOptions.map((b) => (
                      <option key={b.value} value={b.value}>
                        {b.label}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            <div className="border-t pt-4">
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-gray-700">Tanggal Keberangkatan</label>
                  <button
                    type="button"
                    onClick={addKeberangkatan}
                    className="text-sm text-emerald-600 hover:text-emerald-800 font-medium"
                  >
                    + Tambah
                  </button>
                </div>
                {form.tanggal_keberangkatan.length === 0 && (
                  <p className="text-sm text-gray-400">Belum ada tanggal keberangkatan.</p>
                )}
                <div className="space-y-2">
                  {form.tanggal_keberangkatan.map((tk, i) => (
                    <div key={i} className="flex gap-2 items-center">
                      <input
                        type="text"
                        placeholder="Nama (cth: JKG)"
                        value={tk.nama}
                        onChange={(e) => updateKeberangkatan(i, 'nama', e.target.value)}
                        className={inputClass + ' w-32'}
                      />
                      <input
                        type="date"
                        value={tk.tanggal}
                        onChange={(e) => updateKeberangkatan(i, 'tanggal', e.target.value)}
                        className={inputClass}
                      />
                      <button
                        type="button"
                        onClick={() => removeKeberangkatan(i)}
                        className="text-red-500 hover:text-red-700 text-sm font-medium px-2"
                      >
                        Hapus
                      </button>
                    </div>
                  ))}
                </div>
              </div>

            <div className="flex gap-2">
              <button
                type="submit"
                disabled={saving}
                className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
              >
                {saving ? 'Menyimpan...' : editingId ? 'Simpan' : 'Tambah'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300"
              >
                Batal
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <p className="text-gray-500">Memuat data...</p>
      ) : paketList.length === 0 ? (
        <p className="text-gray-500">Tidak ada data paket.</p>
      ) : (
        <div className="overflow-x-auto bg-white rounded-lg shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase w-12">
                  No
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Label
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Tipe
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Tahun
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Bulan
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Keberangkatan
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {paketList.map((p, idx) => (
                <tr key={p.id} className="hover:bg-gray-50">
                  <td className="px-4 py-4 whitespace-nowrap text-sm text-gray-400">
                    {idx + 1}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {p.label}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 capitalize">
                    {p.tipe}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {p.tahun}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {p.bulan ? bulanOptions.find((b) => b.value === p.bulan)?.label || '-' : '-'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {p.tanggal_keberangkatan && p.tanggal_keberangkatan.length > 0
                      ? p.tanggal_keberangkatan.map((tk, i) => (
                          <span key={i} className="inline-block mr-3">
                            <span className="font-medium text-gray-700">{tk.nama}</span>
                            {tk.tanggal && (
                              <span className="text-gray-400"> {formatDisplayDate(tk.tanggal)}</span>
                            )}
                          </span>
                        ))
                      : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-2">
                    <button
                      onClick={() => handleEdit(p)}
                      className="text-blue-600 hover:text-blue-800 font-medium"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDelete(p.id, p.label)}
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
      )}
    </div>
  );
}
