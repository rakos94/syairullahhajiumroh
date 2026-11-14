import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { fetchPaketList, fetchTagihan, downloadTagihanExcel } from '../api';
import StatusBadge from '../components/StatusBadge';
import { formatRupiah } from '../utils/format';

const selectClass =
  'px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500';

function Stat({ label, value, tone = 'gray' }) {
  const tones = {
    gray: 'bg-white',
    amber: 'bg-amber-50 border-amber-200',
    emerald: 'bg-emerald-50 border-emerald-200',
  };
  return (
    <div className={`rounded-lg border p-4 ${tones[tone]}`}>
      <p className="text-sm text-gray-500">{label}</p>
      <p className="mt-1 text-xl font-semibold text-gray-900 tabular-nums">{value}</p>
    </div>
  );
}

export default function Tagihan() {
  const navigate = useNavigate();
  const [paketOptions, setPaketOptions] = useState([]);
  const [paketId, setPaketId] = useState('');
  const [belumLunas, setBelumLunas] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState('');
  const [downloading, setDownloading] = useState(false);

  useEffect(() => {
    fetchPaketList().then(setPaketOptions).catch(() => {});
  }, []);

  useEffect(() => {
    let cancelled = false;
    fetchTagihan({ paket_id: paketId, belum_lunas: belumLunas })
      .then((d) => {
        if (cancelled) return;
        setData(d);
        setError('');
      })
      .catch((err) => !cancelled && setError(err.message));
    return () => {
      cancelled = true;
    };
  }, [paketId, belumLunas]);

  const download = async () => {
    setDownloading(true);
    try {
      await downloadTagihanExcel({ paket_id: paketId, belum_lunas: belumLunas });
    } catch (err) {
      setError(err.message);
    } finally {
      setDownloading(false);
    }
  };

  const s = data?.summary;
  const unpriced = paketOptions.filter((p) => !p.harga && (!paketId || p.id === paketId));

  return (
    <div>
      <div className="sm:flex sm:items-center sm:justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Tagihan</h1>
        <button
          onClick={download}
          disabled={downloading}
          className="mt-3 sm:mt-0 px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-50 disabled:opacity-50"
        >
          {downloading ? 'Mengunduh...' : 'Unduh Excel'}
        </button>
      </div>

      <div className="flex flex-col sm:flex-row gap-3 mb-4">
        <select value={paketId} onChange={(e) => setPaketId(e.target.value)} aria-label="Filter paket" className={selectClass}>
          <option value="">Semua Paket</option>
          {paketOptions.map((p) => (
            <option key={p.id} value={p.id}>
              {p.label}
            </option>
          ))}
        </select>
        <select
          value={belumLunas ? 'belum' : 'semua'}
          onChange={(e) => setBelumLunas(e.target.value === 'belum')}
          aria-label="Filter status tagihan"
          className={selectClass}
        >
          <option value="belum">Masih ada sisa tagihan</option>
          <option value="semua">Semua jamaah</option>
        </select>
      </div>

      {unpriced.length > 0 && (
        <div className="mb-4 p-3 bg-amber-50 border border-amber-200 text-amber-800 rounded-md text-sm">
          Harga belum diatur untuk: {unpriced.map((p) => p.label).join(', ')}. Sisa tagihan jamaah di paket
          tersebut belum bisa dihitung —{' '}
          <Link to="/paket" className="font-medium underline">
            atur harga di menu Paket
          </Link>
          .
        </div>
      )}

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      {!data ? (
        <p className="text-gray-500">Memuat data...</p>
      ) : (
        <>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
            <Stat label="Total tagihan" value={formatRupiah(s.total_tagihan)} />
            <Stat label="Sudah dibayar" value={formatRupiah(s.total_dibayar)} tone="emerald" />
            <Stat label="Sisa tagihan" value={formatRupiah(s.total_sisa)} tone="amber" />
            <Stat label="Jamaah belum lunas" value={`${s.jumlah_belum_lunas} orang`} />
          </div>

          {data.data.length === 0 ? (
            <p className="text-gray-500">
              {belumLunas ? 'Tidak ada jamaah dengan sisa tagihan.' : 'Tidak ada data jamaah.'}
            </p>
          ) : (
            <div className="overflow-x-auto bg-white rounded-lg shadow">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr className="text-xs font-medium text-gray-500 uppercase">
                    <th className="px-4 py-3 text-left">Nama</th>
                    <th className="px-4 py-3 text-left">Paket</th>
                    <th className="px-4 py-3 text-left">Status</th>
                    <th className="px-4 py-3 text-right">Harga</th>
                    <th className="px-4 py-3 text-right">Dibayar</th>
                    <th className="px-4 py-3 text-right">Sisa</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {data.data.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50 cursor-pointer" onClick={() => navigate(`/jamaah/${r.id}`)}>
                      <td className="px-4 py-3 text-sm">
                        <div className="font-medium text-gray-900">{r.nama}</div>
                        <div className="text-xs text-gray-500">{r.no_hp || r.nik}</div>
                        {r.peringatan && <div className="text-xs text-amber-700 mt-0.5">⚠ {r.peringatan}</div>}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">{r.paket}</td>
                      <td className="px-4 py-3">
                        <StatusBadge status={r.status_pembayaran} />
                      </td>
                      <td className="px-4 py-3 text-sm text-right tabular-nums whitespace-nowrap">
                        {r.harga_diatur ? formatRupiah(r.harga) : <span className="text-gray-300">-</span>}
                      </td>
                      <td className="px-4 py-3 text-sm text-right tabular-nums whitespace-nowrap">
                        {formatRupiah(r.total_dibayar)}
                      </td>
                      <td className="px-4 py-3 text-sm text-right tabular-nums whitespace-nowrap font-medium">
                        {!r.harga_diatur ? (
                          <span className="text-gray-300">-</span>
                        ) : r.sisa > 0 ? (
                          formatRupiah(r.sisa)
                        ) : (
                          <span className="text-emerald-700">Lunas</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
