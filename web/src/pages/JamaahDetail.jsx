import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { fetchJamaahById, uploadDocument, getDocumentUrl, getDocumentDownloadUrl, getMultiDocumentUrl, getMultiDocumentDownloadUrl, deleteDocument } from '../api';
import StatusBadge from '../components/StatusBadge';

const singleDocTypes = [
  { key: 'ktp', label: 'KTP', field: 'foto_ktp' },
  { key: 'kk', label: 'KK', field: 'foto_kk' },
  { key: 'paspor', label: 'Paspor', field: 'foto_paspor' },
  { key: 'pasfoto', label: 'Pas Foto', field: 'pasfoto' },
  { key: 'koper_diterima', label: 'Koper Diterima', field: 'foto_koper_diterima' },
];

const multiDocTypes = [
  { key: 'bukti_dp', label: 'Bukti Pembayaran DP', field: 'bukti_dp' },
  { key: 'bukti_pelunasan', label: 'Bukti Pelunasan', field: 'bukti_pelunasan' },
];

function formatDate(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '-';
  const d = new Date(dateStr);
  const day = String(d.getDate()).padStart(2, '0');
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
}

export default function JamaahDetail() {
  const { id } = useParams();
  const [jamaah, setJamaah] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [uploadMsg, setUploadMsg] = useState('');
  const [lightbox, setLightbox] = useState(null);
  const [cacheKey, setCacheKey] = useState(Date.now());

  const loadData = async () => {
    try {
      const data = await fetchJamaahById(id);
      setJamaah(data);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [id]);

  const handleUpload = async (docType, file, nominal) => {
    try {
      setUploadMsg('');
      const result = await uploadDocument(id, docType, file, nominal);
      setUploadMsg(result.message);
      setCacheKey(Date.now());
      loadData();
    } catch (err) {
      setUploadMsg(err.message);
    }
  };

  const handleDeleteDoc = async (docType, index) => {
    if (!confirm('Hapus dokumen ini?')) return;
    try {
      setUploadMsg('');
      const result = await deleteDocument(id, docType, index);
      setUploadMsg(result.message);
      setCacheKey(Date.now());
      loadData();
    } catch (err) {
      setUploadMsg(err.message);
    }
  };

  if (loading) return <p className="text-gray-500">Memuat data...</p>;
  if (error)
    return <p className="text-red-600">{error}</p>;
  if (!jamaah) return null;

  const infoRows = [
    ['Nama', jamaah.nama],
    ['NIK', jamaah.nik],
    ['Nomor Paspor', jamaah.nomor_paspor || '-'],
    ['Alamat', jamaah.alamat || '-'],
    ['No HP', jamaah.no_hp || '-'],
    ['Tempat Lahir', jamaah.tempat_lahir || '-'],
    ['Tanggal Lahir', formatDate(jamaah.tanggal_lahir)],
    ['Jenis Kelamin', jamaah.jenis_kelamin === 'laki-laki' ? 'Laki-laki' : 'Perempuan'],
    ['Paket', jamaah.paket?.label || '-'],
    ['Tanggal Keberangkatan', jamaah.tanggal_keberangkatan
      ? (jamaah.tanggal_keberangkatan.nama
          ? `${jamaah.tanggal_keberangkatan.nama}${jamaah.tanggal_keberangkatan.tanggal ? ` - ${formatDate(jamaah.tanggal_keberangkatan.tanggal)}` : ''}`
          : formatDate(jamaah.tanggal_keberangkatan.tanggal))
      : '-'],
    ['No Rekening Haji', jamaah.no_rekening_haji || '-'],
    ['Tipe Bank', jamaah.tipe_bank || '-'],
    ['Keterangan', jamaah.keterangan || '-'],
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Detail Jamaah</h1>
        <div className="flex gap-2">
          <Link
            to={`/jamaah/${id}/edit`}
            className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
          >
            Edit
          </Link>
          <Link
            to="/"
            className="px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300"
          >
            Kembali
          </Link>
        </div>
      </div>

      {/* Info */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex items-center gap-3 mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Informasi</h2>
          <StatusBadge status={jamaah.status_pembayaran} />
        </div>
        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3">
          {infoRows.map(([label, value]) => (
            <div key={label}>
              <dt className="text-sm text-gray-500">{label}</dt>
              <dd className="text-sm font-medium text-gray-900">{value}</dd>
            </div>
          ))}
        </dl>

        <div className="mt-4 flex flex-wrap gap-4">
          <CheckItem
            label="Batik Nasional Sudah Dijahit"
            checked={jamaah.batik_nasional_sudah_dijahit}
          />
          <CheckItem
            label="Batik KBIH Sudah Diterima"
            checked={jamaah.batik_kbih_sudah_diterima}
          />
          <CheckItem
            label="Koper Sudah Diterima"
            checked={jamaah.koper_sudah_diterima}
          />
        </div>
      </div>

      {/* Documents */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Dokumen</h2>

        {uploadMsg && (
          <div className="mb-4 p-3 bg-emerald-50 text-emerald-700 rounded-md text-sm">
            {uploadMsg}
          </div>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {singleDocTypes.map((doc) => {
            const filePath = jamaah[doc.field];
            const hasFile = Boolean(filePath);
            const isImage = hasFile && /\.(jpg|jpeg|png)$/i.test(filePath);
            const docUrl = getDocumentUrl(id, doc.key) + `?_t=${cacheKey}`;
            return (
              <div
                key={doc.key}
                className="border border-gray-200 rounded-lg p-4"
              >
                <p className="text-sm font-medium text-gray-700 mb-2">
                  {doc.label}
                </p>
                {hasFile ? (
                  <div className="space-y-2">
                    {isImage ? (
                      <button
                        type="button"
                        onClick={() => setLightbox({ url: docUrl, label: doc.label })}
                        className="w-full cursor-pointer"
                      >
                        <img
                          src={docUrl}
                          alt={doc.label}
                          className="w-full max-h-48 object-contain rounded border border-gray-200"
                        />
                      </button>
                    ) : (
                      <a
                        href={docUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="flex items-center justify-center w-full h-40 bg-gray-50 rounded border border-gray-200 text-emerald-600 hover:text-emerald-800 text-sm font-medium"
                      >
                        Lihat PDF
                      </a>
                    )}
                  </div>
                ) : (
                  <div className="flex items-center justify-center w-full h-40 bg-gray-50 rounded border border-dashed border-gray-300">
                    <p className="text-xs text-gray-400">Belum diupload</p>
                  </div>
                )}
                <div className="mt-2 flex gap-2">
                  <label className="inline-flex items-center px-3 py-1.5 bg-gray-100 text-gray-700 text-xs font-medium rounded-md hover:bg-gray-200 cursor-pointer">
                    {hasFile ? 'Ganti File' : 'Upload'}
                    <input
                      type="file"
                      accept=".jpg,.jpeg,.png,.pdf"
                      className="hidden"
                      onChange={(e) => {
                        if (e.target.files[0]) {
                          handleUpload(doc.key, e.target.files[0]);
                          e.target.value = '';
                        }
                      }}
                    />
                  </label>
                  {hasFile && (
                    <a
                      href={getDocumentDownloadUrl(id, doc.key)}
                      className="inline-flex items-center px-3 py-1.5 bg-emerald-50 text-emerald-700 text-xs font-medium rounded-md hover:bg-emerald-100"
                    >
                      Download
                    </a>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {multiDocTypes.map((doc) => {
          const entries = jamaah[doc.field] || [];
          return (
            <div key={doc.key} className="mt-6">
              <div className="flex items-center justify-between mb-3">
                <p className="text-sm font-medium text-gray-700">{doc.label}</p>
                <label className="inline-flex items-center px-3 py-1.5 bg-emerald-50 text-emerald-700 text-xs font-medium rounded-md hover:bg-emerald-100 cursor-pointer">
                  + Tambah
                  <input
                    type="file"
                    accept=".jpg,.jpeg,.png,.pdf"
                    className="hidden"
                    onChange={(e) => {
                      if (e.target.files[0]) {
                        const file = e.target.files[0];
                        e.target.value = '';
                        const nominal = prompt('Masukkan nominal pembayaran (Rp):');
                        if (nominal === null) return;
                        handleUpload(doc.key, file, Number(nominal.replace(/\D/g, '')) || 0);
                      }
                    }}
                  />
                </label>
              </div>
              {entries.length === 0 ? (
                <div className="flex items-center justify-center w-full h-24 bg-gray-50 rounded border border-dashed border-gray-300">
                  <p className="text-xs text-gray-400">Belum ada dokumen</p>
                </div>
              ) : (
                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
                  {entries.map((entry, idx) => {
                    const isImage = /\.(jpg|jpeg|png)$/i.test(entry.file);
                    const fileUrl = getMultiDocumentUrl(id, doc.key, idx) + `&_t=${cacheKey}`;
                    return (
                      <div key={idx} className="relative border border-gray-200 rounded-lg p-2 group">
                        {isImage ? (
                          <button
                            type="button"
                            onClick={() => setLightbox({ url: fileUrl, label: `${doc.label} ${idx + 1}` })}
                            className="w-full cursor-pointer"
                          >
                            <img
                              src={fileUrl}
                              alt={`${doc.label} ${idx + 1}`}
                              className="w-full h-32 object-contain rounded"
                            />
                          </button>
                        ) : (
                          <a
                            href={fileUrl}
                            target="_blank"
                            rel="noreferrer"
                            className="flex items-center justify-center w-full h-32 bg-gray-50 rounded text-emerald-600 hover:text-emerald-800 text-sm font-medium"
                          >
                            Lihat PDF
                          </a>
                        )}
                        <p className="mt-1 text-xs font-medium text-gray-700 text-center">
                          Rp {entry.nominal?.toLocaleString('id-ID') || '0'}
                        </p>
                        <a
                          href={getMultiDocumentDownloadUrl(id, doc.key, idx)}
                          className="mt-1 flex items-center justify-center px-2 py-1 bg-emerald-50 text-emerald-700 text-xs font-medium rounded hover:bg-emerald-100"
                        >
                          Download
                        </a>
                        <button
                          type="button"
                          onClick={() => handleDeleteDoc(doc.key, idx)}
                          className="absolute top-1 right-1 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs hover:bg-red-600 opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                          x
                        </button>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Lightbox Modal */}
      {lightbox && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
          onClick={() => setLightbox(null)}
        >
          <div
            className="relative max-w-4xl max-h-[90vh] p-2"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setLightbox(null)}
              className="absolute -top-2 -right-2 bg-white rounded-full w-8 h-8 flex items-center justify-center text-gray-700 hover:bg-gray-100 shadow-lg text-lg font-bold"
            >
              x
            </button>
            <img
              src={lightbox.url}
              alt={lightbox.label}
              className="max-w-full max-h-[85vh] object-contain rounded-lg"
            />
          </div>
        </div>
      )}
    </div>
  );
}

function CheckItem({ label, checked }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-sm text-gray-700">
      <span
        className={`h-4 w-4 rounded flex items-center justify-center text-white text-xs ${
          checked ? 'bg-emerald-500' : 'bg-gray-300'
        }`}
      >
        {checked ? '✓' : ''}
      </span>
      {label}
    </span>
  );
}
