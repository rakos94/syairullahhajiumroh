import { useState, useEffect } from 'react';
import { fetchAuditLogs } from '../api';

const actionLabels = {
  create: 'Tambah',
  update: 'Ubah',
  delete: 'Hapus',
  upload: 'Upload',
  delete_document: 'Hapus Dokumen',
};

const actionColors = {
  create: 'bg-green-100 text-green-800',
  update: 'bg-blue-100 text-blue-800',
  delete: 'bg-red-100 text-red-800',
  upload: 'bg-purple-100 text-purple-800',
  delete_document: 'bg-orange-100 text-orange-800',
};

const entityTypeLabels = {
  jamaah: 'Jamaah',
  paket: 'Paket',
};

function formatDate(dateStr) {
  const d = new Date(dateStr);
  const pad = (n) => String(n).padStart(2, '0');
  return `${pad(d.getDate())}/${pad(d.getMonth() + 1)}/${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function ChangesDetail({ changes }) {
  if (!changes || changes.length === 0) return null;
  return (
    <div className="mt-2 space-y-1">
      {changes.map((ch, i) => (
        <div key={i} className="text-xs text-gray-600 flex flex-wrap items-baseline gap-1">
          <span className="font-medium text-gray-700">{ch.field}:</span>
          <span className="line-through text-red-500">{ch.old_value || '(kosong)'}</span>
          <span className="text-gray-400">&rarr;</span>
          <span className="text-green-600">{ch.new_value || '(kosong)'}</span>
        </div>
      ))}
    </div>
  );
}

export default function AuditLog() {
  const [logs, setLogs] = useState([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [entityType, setEntityType] = useState('');
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState(null);

  useEffect(() => {
    setLoading(true);
    fetchAuditLogs({ entity_type: entityType, page, limit: 20 })
      .then((data) => {
        setLogs(data.data);
        setTotalPages(data.total_pages);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [page, entityType]);

  const handleFilterChange = (e) => {
    setEntityType(e.target.value);
    setPage(1);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Riwayat Aktivitas</h1>
        <select
          value={entityType}
          onChange={handleFilterChange}
          className="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
        >
          <option value="">Semua</option>
          <option value="jamaah">Jamaah</option>
          <option value="paket">Paket</option>
        </select>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Memuat...</div>
      ) : logs.length === 0 ? (
        <div className="text-center py-12 text-gray-500">Belum ada riwayat aktivitas</div>
      ) : (
        <div className="bg-white shadow rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Waktu</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Admin</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Aksi</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tipe</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Keterangan</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {logs.map((log) => (
                  <tr
                    key={log.id}
                    className={`hover:bg-gray-50 ${log.changes?.length ? 'cursor-pointer' : ''}`}
                    onClick={() => log.changes?.length && setExpandedId(expandedId === log.id ? null : log.id)}
                  >
                    <td className="px-4 py-3 text-sm text-gray-600 whitespace-nowrap align-top">
                      {formatDate(log.created_at)}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-900 font-medium align-top">
                      {log.admin_username}
                    </td>
                    <td className="px-4 py-3 align-top">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${actionColors[log.action] || 'bg-gray-100 text-gray-800'}`}>
                        {actionLabels[log.action] || log.action}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 align-top">
                      {entityTypeLabels[log.entity_type] || log.entity_type}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700 align-top">
                      <div className="flex items-center gap-2">
                        <span>{log.description}</span>
                        {log.changes?.length > 0 && (
                          <span className="text-xs text-gray-400">
                            ({log.changes.length} perubahan)
                          </span>
                        )}
                      </div>
                      {expandedId === log.id && <ChangesDetail changes={log.changes} />}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 text-sm border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
              >
                Sebelumnya
              </button>
              <span className="text-sm text-gray-600">
                Halaman {page} dari {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1.5 text-sm border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
              >
                Selanjutnya
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
