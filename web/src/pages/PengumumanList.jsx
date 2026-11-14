import { useEffect, useState } from 'react';
import { fetchPengumumanList, createPengumuman, updatePengumuman, deletePengumuman } from '../api';

const emptyForm = { judul: '', isi: '', dipublikasikan: true, disematkan: false };

function formatTanggal(dateStr) {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
}

export default function PengumumanList() {
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);

  const loadData = async () => {
    try {
      setLoading(true);
      setList(await fetchPengumumanList());
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

  const handleEdit = (item) => {
    setForm({
      judul: item.judul,
      isi: item.isi,
      dipublikasikan: item.dipublikasikan,
      disematkan: item.disematkan,
    });
    setEditingId(item.id);
    setShowForm(true);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setForm((prev) => ({ ...prev, [name]: type === 'checkbox' ? checked : value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      if (editingId) await updatePengumuman(editingId, form);
      else await createPengumuman(form);
      resetForm();
      loadData();
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const quickToggle = async (item, field) => {
    try {
      await updatePengumuman(item.id, { ...item, [field]: !item[field] });
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDelete = async (item) => {
    if (!confirm(`Hapus pengumuman "${item.judul}"?`)) return;
    try {
      await deletePengumuman(item.id);
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const inputClass =
    'mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500';
  const checkboxClass = 'h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500';

  return (
    <div>
      <div className="sm:flex sm:items-center sm:justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Pengumuman</h1>
          <p className="text-sm text-gray-500 mt-1">
            Pengumuman yang dipublikasikan tampil di halaman depan untuk jamaah.{' '}
            <a href="/" target="_blank" rel="noreferrer" className="text-emerald-700 hover:underline">
              Lihat halaman depan ↗
            </a>
          </p>
        </div>
        {!showForm && (
          <button
            onClick={() => setShowForm(true)}
            className="mt-3 sm:mt-0 inline-flex items-center px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700"
          >
            + Tambah Pengumuman
          </button>
        )}
      </div>

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      {showForm && (
        <form onSubmit={handleSubmit} className="bg-white shadow rounded-lg p-6 mb-6 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {editingId ? 'Edit Pengumuman' : 'Pengumuman Baru'}
          </h2>
          <div>
            <label htmlFor="judul" className="block text-sm font-medium text-gray-700">
              Judul <span className="text-red-500">*</span>
            </label>
            <input
              id="judul"
              name="judul"
              value={form.judul}
              onChange={handleChange}
              maxLength={200}
              required
              className={inputClass}
            />
          </div>
          <div>
            <label htmlFor="isi" className="block text-sm font-medium text-gray-700">
              Isi <span className="text-red-500">*</span>
            </label>
            <textarea
              id="isi"
              name="isi"
              value={form.isi}
              onChange={handleChange}
              rows={6}
              maxLength={5000}
              required
              className={inputClass}
            />
            <p className="mt-1 text-xs text-gray-500">Teks biasa. Baris baru akan ditampilkan apa adanya.</p>
          </div>
          <div className="flex flex-wrap gap-6">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                name="dipublikasikan"
                checked={form.dipublikasikan}
                onChange={handleChange}
                className={checkboxClass}
              />
              Publikasikan (tampil di halaman depan)
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                name="disematkan"
                checked={form.disematkan}
                onChange={handleChange}
                className={checkboxClass}
              />
              Sematkan di atas (penting)
            </label>
          </div>
          <div className="flex gap-3">
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
            >
              {saving ? 'Menyimpan...' : 'Simpan'}
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
      )}

      {loading ? (
        <p className="text-gray-500">Memuat data...</p>
      ) : list.length === 0 ? (
        <p className="text-gray-500">Belum ada pengumuman.</p>
      ) : (
        <div className="space-y-3">
          {list.map((item) => (
            <div key={item.id} className="bg-white shadow rounded-lg p-5">
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-semibold text-gray-900">{item.judul}</h3>
                    {item.dipublikasikan ? (
                      <span className="text-xs font-medium text-emerald-800 bg-emerald-100 rounded-full px-2 py-0.5">
                        Dipublikasikan
                      </span>
                    ) : (
                      <span className="text-xs font-medium text-gray-700 bg-gray-100 rounded-full px-2 py-0.5">
                        Draft
                      </span>
                    )}
                    {item.disematkan && (
                      <span className="text-xs font-medium text-amber-800 bg-amber-100 rounded-full px-2 py-0.5">
                        Disematkan
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    {formatTanggal(item.created_at)}
                    {item.created_by ? ` · oleh ${item.created_by}` : ''}
                  </p>
                  <p className="mt-2 text-sm text-gray-700 whitespace-pre-line break-words line-clamp-3">
                    {item.isi}
                  </p>
                </div>
                <div className="flex flex-wrap sm:flex-col sm:items-end gap-x-3 gap-y-1 text-sm shrink-0">
                  <button
                    onClick={() => quickToggle(item, 'dipublikasikan')}
                    className="text-emerald-700 hover:text-emerald-900 font-medium"
                  >
                    {item.dipublikasikan ? 'Jadikan Draft' : 'Publikasikan'}
                  </button>
                  <button
                    onClick={() => quickToggle(item, 'disematkan')}
                    className="text-amber-700 hover:text-amber-900 font-medium"
                  >
                    {item.disematkan ? 'Lepas Sematan' : 'Sematkan'}
                  </button>
                  <button onClick={() => handleEdit(item)} className="text-blue-600 hover:text-blue-800 font-medium">
                    Edit
                  </button>
                  <button onClick={() => handleDelete(item)} className="text-red-600 hover:text-red-800 font-medium">
                    Hapus
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
