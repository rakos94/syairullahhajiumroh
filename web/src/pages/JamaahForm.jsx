import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { createJamaah, fetchJamaahById, updateJamaah, fetchPaketList } from '../api';

const emptyForm = {
  nama: '',
  nik: '',
  nomor_paspor: '',
  alamat: '',
  no_hp: '',
  tempat_lahir: '',
  tanggal_lahir: '',
  jenis_kelamin: 'laki-laki',
  paket_id: '',
  tanggal_keberangkatan: null,
  status_pembayaran: 'belum_bayar',
  no_rekening_haji: '',
  tipe_bank: '',
  batik_nasional_sudah_dijahit: false,
  batik_kbih_sudah_diterima: false,
  koper_sudah_diterima: false,
  keterangan: '',
};

// Convert ISO date string to yyyy-mm-dd for <input type="date">
function isoToDateInput(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '';
  return dateStr.slice(0, 10);
}

// Convert yyyy-mm-dd to ISO string
function dateInputToIso(val) {
  if (!val) return null;
  const date = new Date(`${val}T00:00:00Z`);
  if (isNaN(date.getTime())) return null;
  return date.toISOString();
}

export default function JamaahForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = Boolean(id);

  const [form, setForm] = useState(emptyForm);
  const [paketOptions, setPaketOptions] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchPaketList().then(setPaketOptions).catch(() => {});
  }, []);

  useEffect(() => {
    if (isEdit) {
      setLoading(true);
      fetchJamaahById(id)
        .then((data) => {
          setForm({
            nama: data.nama || '',
            nik: data.nik || '',
            nomor_paspor: data.nomor_paspor || '',
            alamat: data.alamat || '',
            no_hp: data.no_hp || '',
            tempat_lahir: data.tempat_lahir || '',
            tanggal_lahir: isoToDateInput(data.tanggal_lahir),
            jenis_kelamin: data.jenis_kelamin || 'laki-laki',
            paket_id: data.paket_id || '',
            tanggal_keberangkatan: data.tanggal_keberangkatan || null,
            status_pembayaran: data.status_pembayaran || 'belum_bayar',
            no_rekening_haji: data.no_rekening_haji || '',
            tipe_bank: data.tipe_bank || '',
            batik_nasional_sudah_dijahit: data.batik_nasional_sudah_dijahit || false,
            batik_kbih_sudah_diterima: data.batik_kbih_sudah_diterima || false,
            koper_sudah_diterima: data.koper_sudah_diterima || false,
            keterangan: data.keterangan || '',
          });
        })
        .catch((err) => setError(err.message))
        .finally(() => setLoading(false));
    }
  }, [id, isEdit]);

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setForm((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const payload = {
      ...form,
      tanggal_lahir: dateInputToIso(form.tanggal_lahir) || undefined,
      tanggal_keberangkatan: form.tanggal_keberangkatan || undefined,
    };

    try {
      if (isEdit) {
        await updateJamaah(id, payload);
      } else {
        await createJamaah(payload);
      }
      navigate('/');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const selectedPaket = paketOptions.find((p) => p.id === form.paket_id);
  const paketDepartures = selectedPaket?.tanggal_keberangkatan || [];

  const formatDmyFromIso = (isoStr) => {
    if (!isoStr) return '';
    const d = isoStr.slice(0, 10);
    const [y, m, day] = d.split('-');
    return `${day}/${m}/${y}`;
  };

  const inputClass =
    'mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500';
  const labelClass = 'block text-sm font-medium text-gray-700';

  if (loading && isEdit && !form.nama) {
    return <p className="text-gray-500">Memuat data...</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">
        {isEdit ? 'Edit Jamaah' : 'Tambah Jamaah Baru'}
      </h1>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        onKeyDown={(e) => { if (e.key === 'Enter' && e.target.tagName !== 'TEXTAREA') e.preventDefault(); }}
        className="bg-white shadow rounded-lg p-6"
      >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className={labelClass}>
              Nama <span className="text-red-500">*</span>
            </label>
            <input
              name="nama"
              value={form.nama}
              onChange={handleChange}
              required
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>
              NIK <span className="text-red-500">*</span>
            </label>
            <input
              name="nik"
              value={form.nik}
              onChange={handleChange}
              required
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>Nomor Paspor</label>
            <input
              name="nomor_paspor"
              value={form.nomor_paspor}
              onChange={handleChange}
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>No HP</label>
            <input
              name="no_hp"
              value={form.no_hp}
              onChange={handleChange}
              className={inputClass}
            />
          </div>

          <div className="md:col-span-2">
            <label className={labelClass}>Alamat</label>
            <textarea
              name="alamat"
              value={form.alamat}
              onChange={handleChange}
              rows={2}
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>Tempat Lahir</label>
            <input
              name="tempat_lahir"
              value={form.tempat_lahir}
              onChange={handleChange}
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>Tanggal Lahir</label>
            <input
              type="date"
              name="tanggal_lahir"
              value={form.tanggal_lahir}
              onChange={handleChange}
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>
              Jenis Kelamin <span className="text-red-500">*</span>
            </label>
            <select
              name="jenis_kelamin"
              value={form.jenis_kelamin}
              onChange={handleChange}
              className={inputClass}
            >
              <option value="laki-laki">Laki-laki</option>
              <option value="perempuan">Perempuan</option>
            </select>
          </div>

          <div>
            <label className={labelClass}>
              Paket <span className="text-red-500">*</span>
            </label>
            <select
              name="paket_id"
              value={form.paket_id}
              onChange={(e) => setForm((prev) => ({ ...prev, paket_id: e.target.value, tanggal_keberangkatan: null }))}
              required
              className={inputClass}
            >
              <option value="">-- Pilih Paket --</option>
              {paketOptions.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className={labelClass}>Tanggal Keberangkatan</label>
            {paketDepartures.length > 0 ? (
              <select
                value={form.tanggal_keberangkatan ? form.tanggal_keberangkatan.nama : ''}
                onChange={(e) => {
                  const tk = paketDepartures.find((d) => d.nama === e.target.value);
                  setForm((prev) => ({
                    ...prev,
                    tanggal_keberangkatan: tk ? { nama: tk.nama, tanggal: tk.tanggal } : null,
                  }));
                }}
                className={inputClass}
              >
                <option value="">-- Pilih Keberangkatan --</option>
                {paketDepartures.map((tk, i) => (
                  <option key={i} value={tk.nama}>
                    {tk.nama}{tk.tanggal ? ` - ${formatDmyFromIso(tk.tanggal)}` : ''}
                  </option>
                ))}
              </select>
            ) : (
              <input
                type="date"
                name="tanggal_keberangkatan_manual"
                value={form.tanggal_keberangkatan ? isoToDateInput(form.tanggal_keberangkatan.tanggal) : ''}
                onChange={(e) => {
                  const val = e.target.value;
                  const iso = dateInputToIso(val);
                  setForm((prev) => ({
                    ...prev,
                    tanggal_keberangkatan: val ? { tanggal: iso || val } : null,
                  }));
                }}
                className={inputClass}
              />
            )}
          </div>

          <div>
            <label className={labelClass}>
              Status Pembayaran <span className="text-red-500">*</span>
            </label>
            <select
              name="status_pembayaran"
              value={form.status_pembayaran}
              onChange={handleChange}
              className={inputClass}
            >
              <option value="belum_bayar">Belum Bayar</option>
              <option value="dp">DP</option>
              <option value="lunas">Lunas</option>
            </select>
          </div>

          <div>
            <label className={labelClass}>No Rekening Haji</label>
            <input
              name="no_rekening_haji"
              value={form.no_rekening_haji}
              onChange={handleChange}
              className={inputClass}
            />
          </div>

          <div>
            <label className={labelClass}>Tipe Bank</label>
            <input
              name="tipe_bank"
              value={form.tipe_bank}
              onChange={handleChange}
              placeholder="BRI, BNI, BSI, dll"
              className={inputClass}
            />
          </div>

          <div className="md:col-span-2">
            <label className={labelClass}>Keterangan</label>
            <textarea
              name="keterangan"
              value={form.keterangan}
              onChange={handleChange}
              rows={2}
              className={inputClass}
            />
          </div>

          <div className="md:col-span-2 flex flex-wrap gap-6 pt-2">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                name="batik_nasional_sudah_dijahit"
                checked={form.batik_nasional_sudah_dijahit}
                onChange={handleChange}
                className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
              />
              Batik Nasional Sudah Dijahit
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                name="batik_kbih_sudah_diterima"
                checked={form.batik_kbih_sudah_diterima}
                onChange={handleChange}
                className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
              />
              Batik KBIH Sudah Diterima
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                name="koper_sudah_diterima"
                checked={form.koper_sudah_diterima}
                onChange={handleChange}
                className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
              />
              Koper Sudah Diterima
            </label>
          </div>
        </div>

        <div className="mt-6 flex gap-3">
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
          >
            {loading ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Simpan'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/')}
            className="px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300"
          >
            Batal
          </button>
        </div>
      </form>
    </div>
  );
}
