import { useState, useEffect } from 'react';
import { fetchStatistics } from '../api';

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await fetchStatistics();
      setStats(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="text-center py-12 text-gray-500">Memuat statistik...</div>;
  }

  if (error) {
    return (
      <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>
    );
  }

  if (!stats) return null;

  const paymentMap = {};
  (stats.payment_breakdown || []).forEach((p) => {
    paymentMap[p.status] = p.count;
  });

  const allPakets = stats.paket_breakdown || [];
  const hajiPakets = allPakets.filter((p) => p.tipe === 'haji');
  const umrohPakets = allPakets.filter((p) => p.tipe === 'umroh');

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <SummaryCard label="Total Jamaah" value={stats.total} color="emerald" />
        <SummaryCard label="Jamaah Haji" value={stats.total_haji} color="blue" />
        <SummaryCard label="Jamaah Umroh" value={stats.total_umroh} color="purple" />
      </div>

      {/* Payment Status */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <PaymentCard
          label="Belum Bayar"
          count={paymentMap['belum_bayar'] || 0}
          total={stats.total}
          color="red"
        />
        <PaymentCard
          label="DP"
          count={paymentMap['dp'] || 0}
          total={stats.total}
          color="yellow"
        />
        <PaymentCard
          label="Lunas"
          count={paymentMap['lunas'] || 0}
          total={stats.total}
          color="green"
        />
      </div>

      {/* Haji Section */}
      {hajiPakets.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 border-b border-blue-200 pb-2">Haji</h2>

          {/* Kelengkapan Haji */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <ProgressCard label="Batik Nasional" done={stats.batik_nasional_done} total={stats.total_haji_completion} />
            <ProgressCard label="Batik KBIH" done={stats.batik_kbih_done} total={stats.total_haji_completion} />
            <ProgressCard label="Koper Diterima" done={stats.koper_done} total={stats.total_haji_completion} />
          </div>

          {/* Haji Paket Table */}
          <PaketTable pakets={hajiPakets} />
        </div>
      )}

      {/* Umroh Section */}
      {umrohPakets.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 border-b border-purple-200 pb-2">Umroh</h2>
          <PaketTable pakets={umrohPakets} />
        </div>
      )}
    </div>
  );
}

function SummaryCard({ label, value, color }) {
  const colors = {
    emerald: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    purple: 'bg-purple-50 text-purple-700 border-purple-200',
  };
  return (
    <div className={`rounded-lg border p-5 ${colors[color]}`}>
      <p className="text-sm font-medium opacity-75">{label}</p>
      <p className="text-3xl font-bold mt-1">{value}</p>
    </div>
  );
}

function PaymentCard({ label, count, total, color }) {
  const colors = {
    red: 'bg-red-50 text-red-700 border-red-200',
    yellow: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    green: 'bg-green-50 text-green-700 border-green-200',
  };
  const pct = total > 0 ? Math.round((count / total) * 100) : 0;
  return (
    <div className={`rounded-lg border p-5 ${colors[color]}`}>
      <p className="text-sm font-medium opacity-75">{label}</p>
      <p className="text-3xl font-bold mt-1">{count}</p>
      <p className="text-sm opacity-60 mt-1">{pct}% dari total</p>
    </div>
  );
}

function PaketTable({ pakets }) {
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Paket</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Total</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Belum Bayar</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">DP</th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Lunas</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {pakets.map((p) => (
              <tr key={p.paket_id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {p.label || '-'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-right font-semibold text-gray-900">
                  {p.total}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-red-600">
                  {p.belum_bayar}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-yellow-600">
                  {p.dp}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-green-600">
                  {p.lunas}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ProgressCard({ label, done, total }) {
  const pct = total > 0 ? Math.round((done / total) * 100) : 0;
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-5">
      <p className="text-sm font-medium text-gray-600">{label}</p>
      <p className="text-2xl font-bold text-gray-900 mt-1">
        {done} <span className="text-sm font-normal text-gray-400">/ {total}</span>
      </p>
      <div className="mt-3 w-full bg-gray-200 rounded-full h-2">
        <div
          className="bg-emerald-500 h-2 rounded-full transition-all"
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="text-xs text-gray-400 mt-1">{pct}%</p>
    </div>
  );
}
