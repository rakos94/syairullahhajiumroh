import { useState } from 'react';
import { cekKamarManasik } from '../api';
import ManasikQr from './ManasikQr';

function formatDate(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '';
  const d = new Date(dateStr);
  const day = String(d.getUTCDate()).padStart(2, '0');
  const month = String(d.getUTCMonth() + 1).padStart(2, '0');
  return `${day}/${month}/${d.getUTCFullYear()}`;
}

// 0812... -> 62812... for wa.me links
function waNumber(hp) {
  const digits = (hp || '').replace(/\D/g, '');
  if (digits.startsWith('0')) return `62${digits.slice(1)}`;
  return digits;
}

function StatusList({ title, items }) {
  if (!items || items.length === 0) return null;
  const done = items.filter((i) => i.selesai).length;
  return (
    <div>
      <div className="flex justify-between items-baseline">
        <p className="text-xs uppercase tracking-wide text-gray-500">{title}</p>
        <p className="text-xs text-gray-500">
          {done}/{items.length}
        </p>
      </div>
      <ul className="mt-1 space-y-1">
        {items.map((i) => (
          <li key={i.nama} className="flex items-center gap-2 text-sm">
            <span
              aria-hidden="true"
              className={`inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                i.selesai ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
              }`}
            >
              {i.selesai ? '✓' : '!'}
            </span>
            <span className={i.selesai ? 'text-gray-700' : 'text-gray-900 font-medium'}>{i.nama}</span>
            <span className="sr-only">{i.selesai ? 'sudah' : 'belum'}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

function KontakKetua({ label, list }) {
  if (!list || list.length === 0) return null;
  return (
    <div>
      <p className="text-xs uppercase tracking-wide text-gray-500">{label}</p>
      {list.map((k) => (
        <div key={`${k.nama}-${k.no_hp}`} className="mt-1 flex items-center justify-between gap-3">
          <span className="text-sm font-medium text-gray-900">{k.nama}</span>
          {k.no_hp && (
            <span className="flex gap-2 shrink-0">
              <a
                href={`tel:${k.no_hp}`}
                className="px-2.5 py-1 text-xs font-medium rounded-md border border-gray-300 text-gray-700 hover:bg-gray-50"
              >
                Telepon
              </a>
              <a
                href={`https://wa.me/${waNumber(k.no_hp)}`}
                target="_blank"
                rel="noreferrer"
                className="px-2.5 py-1 text-xs font-medium rounded-md bg-emerald-600 text-white hover:bg-emerald-700"
              >
                WhatsApp
              </a>
            </span>
          )}
        </div>
      ))}
    </div>
  );
}

export default function CekKamarCard() {
  const [nik, setNik] = useState('');
  const [tanggalLahir, setTanggalLahir] = useState('');
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setResult(null);
    setLoading(true);
    try {
      setResult(await cekKamarManasik(nik.trim(), tanggalLahir));
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const reset = () => {
    setResult(null);
    setNik('');
    setTanggalLahir('');
  };

  const inputClass =
    'w-full px-3 py-3 border border-gray-300 rounded-md text-base focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500';

  const keberangkatan = result?.tanggal_keberangkatan;
  const keberangkatanText = keberangkatan
    ? [keberangkatan.nama, formatDate(keberangkatan.tanggal)].filter(Boolean).join(' - ')
    : '';
  const sudahDitentukan = result && (result.gedung_manasik || result.kamar_manasik);

  return (
    <>
      {!result && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <p className="text-sm text-gray-600 mb-4">
            Masukkan NIK (sesuai KTP) dan tanggal lahir Anda untuk melihat gedung dan kamar
            manasik di asrama haji.
          </p>

          {error && (
            <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="nik" className="block text-sm font-medium text-gray-700 mb-1">
                NIK
              </label>
              <input
                id="nik"
                inputMode="numeric"
                autoComplete="off"
                value={nik}
                onChange={(e) => setNik(e.target.value.replace(/\D/g, ''))}
                maxLength={16}
                placeholder="16 digit NIK"
                required
                className={inputClass}
              />
            </div>
            <div>
              <label htmlFor="tanggal_lahir" className="block text-sm font-medium text-gray-700 mb-1">
                Tanggal Lahir
              </label>
              <input
                id="tanggal_lahir"
                type="date"
                value={tanggalLahir}
                onChange={(e) => setTanggalLahir(e.target.value)}
                required
                className={inputClass}
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 px-4 bg-emerald-600 text-white font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
            >
              {loading ? 'Mencari...' : 'Cek Kamar'}
            </button>
          </form>
        </div>
      )}

      {result && (
        <div className="bg-white rounded-lg shadow-md p-6">
          <p className="text-sm text-gray-500">Assalamu'alaikum,</p>
          <p className="text-xl font-semibold text-gray-900">{result.nama}</p>
          {result.paket && <p className="text-sm text-gray-600 mt-1">{result.paket}</p>}
          {keberangkatanText && (
            <p className="text-sm text-gray-600">Keberangkatan: {keberangkatanText}</p>
          )}

          {sudahDitentukan ? (
            <div className="mt-6 grid grid-cols-2 gap-3">
              <div className="rounded-lg bg-emerald-50 border border-emerald-200 p-4 text-center">
                <p className="text-xs uppercase tracking-wide text-emerald-700">Gedung</p>
                <p className="text-2xl font-bold text-emerald-900 mt-1 break-words">
                  {result.gedung_manasik || '-'}
                </p>
              </div>
              <div className="rounded-lg bg-emerald-50 border border-emerald-200 p-4 text-center">
                <p className="text-xs uppercase tracking-wide text-emerald-700">Kamar</p>
                <p className="text-2xl font-bold text-emerald-900 mt-1 break-words">
                  {result.kamar_manasik || '-'}
                </p>
              </div>
            </div>
          ) : (
            <div className="mt-6 rounded-lg bg-amber-50 border border-amber-200 p-4 text-sm text-amber-800">
              Gedung dan kamar manasik Anda belum ditentukan. Silakan cek kembali nanti atau
              hubungi petugas.
            </div>
          )}

          {(result.rombongan || result.regu) && (
            <div className="mt-3 grid grid-cols-2 gap-3">
              <div className="rounded-lg bg-gray-50 border border-gray-200 p-3 text-center">
                <p className="text-xs uppercase tracking-wide text-gray-500">Rombongan</p>
                <p className="text-lg font-semibold text-gray-900 break-words">{result.rombongan || '-'}</p>
              </div>
              <div className="rounded-lg bg-gray-50 border border-gray-200 p-3 text-center">
                <p className="text-xs uppercase tracking-wide text-gray-500">Regu</p>
                <p className="text-lg font-semibold text-gray-900 break-words">{result.regu || '-'}</p>
              </div>
            </div>
          )}

          {(result.ketua_regu?.length > 0 || result.ketua_rombongan?.length > 0) && (
            <div className="mt-4 space-y-3 rounded-lg border border-gray-200 p-4">
              <KontakKetua label="Ketua Regu" list={result.ketua_regu} />
              <KontakKetua label="Ketua Rombongan" list={result.ketua_rombongan} />
            </div>
          )}

          {(result.kelengkapan?.length > 0 || result.dokumen?.length > 0) && (
            <div className="mt-4 space-y-4 rounded-lg border border-gray-200 p-4">
              <StatusList title="Kelengkapan" items={result.kelengkapan} />
              <StatusList title="Dokumen diterima kantor" items={result.dokumen} />
              {[...(result.kelengkapan || []), ...(result.dokumen || [])].some((i) => !i.selesai) && (
                <p className="text-xs text-gray-500">
                  Tanda ! berarti belum lengkap. Hubungi kantor Syairullah jika menurut Anda sudah diserahkan.
                </p>
              )}
            </div>
          )}

          {result.qr_token && <ManasikQr token={result.qr_token} nama={result.nama} />}

          <button
            onClick={reset}
            className="mt-6 w-full py-3 px-4 bg-gray-100 text-gray-700 font-medium rounded-md hover:bg-gray-200"
          >
            Cek Jamaah Lain
          </button>
        </div>
      )}
    </>
  );
}
