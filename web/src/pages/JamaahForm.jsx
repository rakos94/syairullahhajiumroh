import { useEffect, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { createJamaah, fetchJamaahById, updateJamaah, fetchPaketList } from '../api';
import RupiahInput from '../components/RupiahInput';
import TanggalInput from '../components/TanggalInput';
import { formatRupiah } from '../utils/format';
import { kelengkapanItems } from '../utils/kelengkapan';

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
  pilihan_harga: '',
  potongan: 0,
  keterangan_potongan: '',
  no_rekening_haji: '',
  tipe_bank: '',
  batik_nasional_sudah_dijahit: false,
  batik_kbih_sudah_diterima: false,
  koper_sudah_diterima: false,
  seragam_umroh_sudah_diterima: false,
  gedung_manasik: '',
  kamar_manasik: '',
  rombongan: '',
  regu: '',
  ketua_regu: false,
  ketua_rombongan: false,
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
  // JS Date rolls invalid days over (1990-02-31 -> 1990-03-03); reject instead
  if (isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== val) return null;
  return date.toISOString();
}

function Section({ title, description, children }) {
  return (
    <section className="bg-white shadow rounded-lg p-6">
      <div className="mb-4 border-b border-gray-100 pb-3">
        <h2 className="text-base font-semibold text-gray-900">{title}</h2>
        {description && <p className="mt-0.5 text-sm text-gray-500">{description}</p>}
      </div>
      {children}
    </section>
  );
}

export default function JamaahForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = Boolean(id);
  // "Daftarkan ke Paket Lain": prefill personal data from an existing registration.
  const [searchParams] = useSearchParams();
  const dariId = isEdit ? '' : searchParams.get('dari') || '';
  const [dariPaket, setDariPaket] = useState('');
  // Other registrations of the same person; their data diri follows this edit.
  const [paketLain, setPaketLain] = useState([]);

  const [form, setForm] = useState(emptyForm);
  const [paketOptions, setPaketOptions] = useState([]);
  const [loading, setLoading] = useState(false);
  // Total of uploaded DP + pelunasan receipts (edit mode); used to preview the
  // automatic payment status.
  const [totalDibayar, setTotalDibayar] = useState(0);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchPaketList().then(setPaketOptions).catch(() => {});
  }, []);

  useEffect(() => {
    if (isEdit) {
      setLoading(true);
      fetchJamaahById(id)
        .then((data) => {
          setTotalDibayar(data.tagihan?.total_dibayar || 0);
          setPaketLain(data.paket_lain || []);
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
            pilihan_harga: data.pilihan_harga || '',
            potongan: data.potongan || 0,
            keterangan_potongan: data.keterangan_potongan || '',
            no_rekening_haji: data.no_rekening_haji || '',
            tipe_bank: data.tipe_bank || '',
            batik_nasional_sudah_dijahit: data.batik_nasional_sudah_dijahit || false,
            batik_kbih_sudah_diterima: data.batik_kbih_sudah_diterima || false,
            koper_sudah_diterima: data.koper_sudah_diterima || false,
            seragam_umroh_sudah_diterima: data.seragam_umroh_sudah_diterima || false,
            gedung_manasik: data.gedung_manasik || '',
            kamar_manasik: data.kamar_manasik || '',
            rombongan: data.rombongan || '',
            regu: data.regu || '',
            ketua_regu: data.ketua_regu || false,
            ketua_rombongan: data.ketua_rombongan || false,
            keterangan: data.keterangan || '',
          });
        })
        .catch((err) => setError(err.message))
        .finally(() => setLoading(false));
    }
  }, [id, isEdit]);

  useEffect(() => {
    if (!dariId) return;
    setLoading(true);
    fetchJamaahById(dariId)
      .then((data) => {
        setDariPaket(data.paket?.label || '');
        // Only the person's own data; paket, payment, manasik and documents
        // belong to each registration.
        setForm({
          ...emptyForm,
          nama: data.nama || '',
          nik: data.nik || '',
          nomor_paspor: data.nomor_paspor || '',
          alamat: data.alamat || '',
          no_hp: data.no_hp || '',
          tempat_lahir: data.tempat_lahir || '',
          tanggal_lahir: isoToDateInput(data.tanggal_lahir),
          jenis_kelamin: data.jenis_kelamin || 'laki-laki',
        });
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [dariId]);

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
      status_pembayaran: statusOtomatis || form.status_pembayaran,
      tanggal_keberangkatan: form.tanggal_keberangkatan || undefined,
    };

    try {
      if (isEdit) {
        await updateJamaah(id, payload);
        navigate(`/jamaah/${id}`);
      } else {
        const created = await createJamaah(payload);
        navigate(`/jamaah/${created.id}`);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const selectedPaket = paketOptions.find((p) => p.id === form.paket_id);
  const paketPilihanHarga = selectedPaket?.pilihan_harga || [];
  const pilihanTerpilih = paketPilihanHarga.find((o) => o.nama === form.pilihan_harga);
  const hargaDasar = pilihanTerpilih ? pilihanTerpilih.harga : selectedPaket?.harga || 0;
  // Same rule as the server: with a price, the status follows the receipts.
  const hargaAkhir = Math.max(hargaDasar - form.potongan, 0);
  const statusOtomatis =
    hargaDasar > 0 ? (totalDibayar <= 0 ? 'belum_bayar' : totalDibayar >= hargaAkhir ? 'lunas' : 'dp') : null;
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

  if (loading && (isEdit || dariId) && !form.nama) {
    return <p className="text-gray-500">Memuat data...</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">
        {isEdit ? 'Edit Jamaah' : dariId ? 'Daftarkan ke Paket Lain' : 'Tambah Jamaah Baru'}
      </h1>

      {dariId && form.nama && (
        <div className="mb-4 p-3 bg-emerald-50 text-emerald-800 rounded-md text-sm">
          Data diri {form.nama} disalin dari pendaftaran{dariPaket ? ` ${dariPaket}` : ''}, dan dokumen
          identitas (KTP, KK, paspor, pas foto) otomatis dipakai bersama. Pilih paket yang baru; tagihan,
          manasik, dan bukti pembayaran dicatat terpisah untuk paket ini.
        </div>
      )}

      {isEdit && paketLain.length > 0 && (
        <div className="mb-4 p-3 bg-blue-50 text-blue-800 rounded-md text-sm">
          Jamaah ini juga terdaftar di {paketLain.map((p) => p.paket).filter(Boolean).join(', ') || `${paketLain.length} paket lain`}.
          Perubahan data diri (nama, NIK, paspor, alamat, no HP, tempat/tanggal lahir, jenis kelamin) ikut
          diperbarui di sana.
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        onKeyDown={(e) => { if (e.key === 'Enter' && e.target.tagName !== 'TEXTAREA') e.preventDefault(); }}
        className="space-y-6"
      >
        <Section title="Data Pribadi" description="Sesuai KTP dan paspor.">
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
              <label className={labelClass}>Nomor Paspor</label>
              <input
                name="nomor_paspor"
                value={form.nomor_paspor}
                onChange={handleChange}
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
              <TanggalInput
                value={form.tanggal_lahir}
                onChange={(v) => setForm((prev) => ({ ...prev, tanggal_lahir: v }))}
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
          </div>
        </Section>

        <Section title="Paket & Keberangkatan">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className={labelClass}>
                Paket <span className="text-red-500">*</span>
              </label>
              <select
                name="paket_id"
                value={form.paket_id}
                onChange={(e) => setForm((prev) => ({ ...prev, paket_id: e.target.value, tanggal_keberangkatan: null, pilihan_harga: '' }))}
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
          </div>
        </Section>

        <Section title="Pembayaran">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label htmlFor="status_pembayaran" className={labelClass}>
                Status Pembayaran <span className="text-red-500">*</span>
              </label>
              <select
                id="status_pembayaran"
                name="status_pembayaran"
                value={statusOtomatis || form.status_pembayaran}
                onChange={handleChange}
                disabled={Boolean(statusOtomatis)}
                aria-describedby="status-pembayaran-hint"
                className={`${inputClass} disabled:bg-gray-50 disabled:text-gray-700`}
              >
                <option value="belum_bayar">Belum Bayar</option>
                <option value="dp">DP</option>
                <option value="lunas">Lunas</option>
              </select>
              <p id="status-pembayaran-hint" className="mt-1 text-xs text-gray-500">
                {statusOtomatis
                  ? 'Otomatis dari bukti DP & pelunasan yang diunggah.'
                  : form.paket_id
                    ? 'Diisi manual karena harga paket belum diatur.'
                    : ''}
              </p>
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
          </div>

          <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
            {paketPilihanHarga.length > 0 && (
              <div>
                <label htmlFor="pilihan_harga" className={labelClass}>
                  Pilihan Harga
                </label>
                <select
                  id="pilihan_harga"
                  name="pilihan_harga"
                  value={form.pilihan_harga}
                  onChange={handleChange}
                  className={inputClass}
                >
                  <option value="">
                    {selectedPaket?.harga ? `Harga paket (${formatRupiah(selectedPaket.harga)})` : '-- Pilih --'}
                  </option>
                  {paketPilihanHarga.map((o) => (
                    <option key={o.nama} value={o.nama}>
                      {o.nama} ({formatRupiah(o.harga)})
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div>
              <label htmlFor="potongan" className={labelClass}>
                Potongan / Diskon
              </label>
              <RupiahInput
                id="potongan"
                value={form.potongan}
                onChange={(potongan) => setForm((prev) => ({ ...prev, potongan }))}
                placeholder="0"
                className={inputClass}
              />
            </div>

            {form.potongan > 0 && (
              <div>
                <label htmlFor="keterangan_potongan" className={labelClass}>
                  Alasan Potongan
                </label>
                <input
                  id="keterangan_potongan"
                  name="keterangan_potongan"
                  value={form.keterangan_potongan}
                  onChange={handleChange}
                  placeholder="Contoh: daftar awal, keluarga"
                  className={inputClass}
                />
              </div>
            )}
          </div>

          <p className="mt-3 text-sm text-gray-600">
            {hargaDasar > 0 ? (
              <>
                Harga untuk jamaah ini:{' '}
                <span className="font-semibold text-gray-900">{formatRupiah(Math.max(hargaDasar - form.potongan, 0))}</span>
                {form.potongan > 0 && (
                  <span className="text-gray-500">
                    {' '}
                    ({formatRupiah(hargaDasar)} − {formatRupiah(form.potongan)})
                  </span>
                )}
                {form.potongan > hargaDasar && (
                  <span className="block text-red-600">Potongan melebihi harga.</span>
                )}
              </>
            ) : form.paket_id ? (
              <span className="text-amber-700">Harga paket belum diatur; sisa tagihan belum bisa dihitung.</span>
            ) : null}
          </p>
        </Section>

        {selectedPaket?.tipe === 'haji' && (
          <Section
            title="Manasik"
            description="Penempatan di asrama haji serta rombongan & regu. Jamaah dapat melihatnya sendiri di halaman Cek Kamar, termasuk nama dan No HP ketua regu/rombongan."
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className={labelClass}>Gedung Manasik</label>
                <input
                  name="gedung_manasik"
                  value={form.gedung_manasik}
                  onChange={handleChange}
                  placeholder="Contoh: Gedung A"
                  className={inputClass}
                />
              </div>

              <div>
                <label className={labelClass}>Kamar Manasik</label>
                <input
                  name="kamar_manasik"
                  value={form.kamar_manasik}
                  onChange={handleChange}
                  placeholder="Contoh: 101"
                  className={inputClass}
                />
              </div>

              <div>
                <label className={labelClass}>Rombongan</label>
                <input
                  name="rombongan"
                  value={form.rombongan}
                  onChange={handleChange}
                  placeholder="Contoh: 1"
                  className={inputClass}
                />
              </div>

              <div>
                <label className={labelClass}>Regu</label>
                <input
                  name="regu"
                  value={form.regu}
                  onChange={handleChange}
                  placeholder="Contoh: 2"
                  className={inputClass}
                />
              </div>

              <div className="md:col-span-2 flex flex-wrap gap-6">
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    name="ketua_regu"
                    checked={form.ketua_regu}
                    onChange={handleChange}
                    className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                  />
                  Ketua Regu
                </label>
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    name="ketua_rombongan"
                    checked={form.ketua_rombongan}
                    onChange={handleChange}
                    className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                  />
                  Ketua Rombongan
                </label>
              </div>
            </div>
          </Section>
        )}

        <Section title="Kelengkapan">
          <div className="flex flex-wrap gap-6">
            {kelengkapanItems(selectedPaket?.tipe).map((item) => (
              <label key={item.field} className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  name={item.field}
                  checked={form[item.field]}
                  onChange={handleChange}
                  className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                />
                {item.label}
              </label>
            ))}
          </div>
        </Section>

        <Section title="Catatan">
          <div>
            <label className={labelClass}>Keterangan</label>
            <textarea
              name="keterangan"
              value={form.keterangan}
              onChange={handleChange}
              rows={2}
              className={inputClass}
            />
          </div>
        </Section>

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
          >
            {loading ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Simpan'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/jamaah')}
            className="px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300"
          >
            Batal
          </button>
        </div>
      </form>
    </div>
  );
}
