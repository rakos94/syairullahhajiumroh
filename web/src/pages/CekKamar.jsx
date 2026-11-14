import { useState } from 'react';
import { Link } from 'react-router-dom';
import { cekKamarManasik } from '../api';

function formatDate(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '';
  const d = new Date(dateStr);
  const day = String(d.getUTCDate()).padStart(2, '0');
  const month = String(d.getUTCMonth() + 1).padStart(2, '0');
  return `${day}/${month}/${d.getUTCFullYear()}`;
}

export default function CekKamar() {
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
    <div className="min-h-screen bg-gray-50 px-4 py-10">
      <div className="max-w-md mx-auto">
        <h1 className="text-2xl font-bold text-center text-emerald-700">Syairullah Haji & Umroh</h1>
        <p className="text-center text-gray-500 mt-1 mb-6">Cek Kamar Manasik</p>

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

            <button
              onClick={reset}
              className="mt-6 w-full py-3 px-4 bg-gray-100 text-gray-700 font-medium rounded-md hover:bg-gray-200"
            >
              Cek Jamaah Lain
            </button>
          </div>
        )}

        <p className="mt-6 text-center text-sm text-gray-400">
          <Link to="/login" className="hover:underline">
            Login admin
          </Link>
        </p>
      </div>
    </div>
  );
}
