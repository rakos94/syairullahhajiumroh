import { useCallback, useEffect, useRef, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { fetchJamaahById, uploadDocument, getDocumentUrl, getDocumentDownloadUrl, getMultiDocumentUrl, getMultiDocumentDownloadUrl, deleteDocument } from '../api';
import { formatRupiah } from '../utils/format';
import { kelengkapanItems } from '../utils/kelengkapan';
import StatusBadge from '../components/StatusBadge';
import CopyButton from '../components/CopyButton';
import NominalDialog from '../components/NominalDialog';

const singleDocTypes = [
  // shared: identity documents are shared by all pakets of the same person.
  { key: 'ktp', label: 'KTP', field: 'foto_ktp', shared: true },
  { key: 'kk', label: 'KK', field: 'foto_kk', shared: true },
  { key: 'paspor', label: 'Paspor', field: 'foto_paspor', shared: true },
  { key: 'pasfoto', label: 'Pas Foto', field: 'pasfoto', shared: true },
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

// Value with a copy button next to it; "-" when empty.
function withCopy(value, label) {
  if (!value) return '-';
  return (
    <span className="inline-flex items-center gap-1">
      <span className="tabular-nums">{value}</span>
      <CopyButton text={value} label={label} />
    </span>
  );
}

export default function JamaahDetail() {
  const { id } = useParams();
  const [jamaah, setJamaah] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [uploadMsg, setUploadMsg] = useState('');
  const [lightbox, setLightbox] = useState(null);
  const [nominalUpload, setNominalUpload] = useState(null); // {docKey, label, file}
  const [cacheKey, setCacheKey] = useState(Date.now());

  const loadedAtRef = useRef(0);

  const loadData = useCallback(async () => {
    try {
      const data = await fetchJamaahById(id);
      loadedAtRef.current = Date.now();
      setJamaah(data);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Document links carry a signature valid for 2 hours (dokumen_akses).
  // Refresh it quietly before it expires so a page left open keeps working.
  // Timed from our own last load, not the server's expiry, so a wrong
  // computer clock can't cause refresh loops or missed refreshes.
  const aksesSig = jamaah?.dokumen_akses?.sig;
  useEffect(() => {
    if (!aksesSig) return undefined;
    const REFRESH_AFTER_MS = 90 * 60 * 1000;
    const timer = setTimeout(loadData, REFRESH_AFTER_MS);
    // Timers can be delayed while a laptop sleeps or the tab is in the
    // background; check again when the admin comes back to the page.
    const onReturn = () => {
      if (document.visibilityState === 'visible' && Date.now() - loadedAtRef.current > REFRESH_AFTER_MS) {
        loadData();
      }
    };
    document.addEventListener('visibilitychange', onReturn);
    window.addEventListener('focus', onReturn);
    return () => {
      clearTimeout(timer);
      document.removeEventListener('visibilitychange', onReturn);
      window.removeEventListener('focus', onReturn);
    };
  }, [aksesSig, loadData]);

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
    ['NIK', withCopy(jamaah.nik, 'Salin NIK')],
    ['Nomor Paspor', withCopy(jamaah.nomor_paspor, 'Salin nomor paspor')],
    ['Alamat', jamaah.alamat || '-'],
    ['No HP', withCopy(jamaah.no_hp, 'Salin No HP')],
    ['Tempat Lahir', jamaah.tempat_lahir || '-'],
    ['Tanggal Lahir', formatDate(jamaah.tanggal_lahir)],
    ['Jenis Kelamin', jamaah.jenis_kelamin === 'laki-laki' ? 'Laki-laki' : 'Perempuan'],
    ['Paket', jamaah.paket?.label || '-'],
    ['Tanggal Keberangkatan', jamaah.tanggal_keberangkatan
      ? (jamaah.tanggal_keberangkatan.nama
          ? `${jamaah.tanggal_keberangkatan.nama}${formatDate(jamaah.tanggal_keberangkatan.tanggal) !== '-' ? ` - ${formatDate(jamaah.tanggal_keberangkatan.tanggal)}` : ''}`
          : formatDate(jamaah.tanggal_keberangkatan.tanggal))
      : '-'],
    ['No Rekening Haji', withCopy(jamaah.no_rekening_haji, 'Salin No Rekening Haji')],
    ['Tipe Bank', jamaah.tipe_bank || '-'],
    ...(jamaah.paket?.tipe === 'haji'
      ? [
          ['Gedung Manasik', jamaah.gedung_manasik || '-'],
          ['Kamar Manasik', jamaah.kamar_manasik || '-'],
          ['Rombongan / Regu', [jamaah.rombongan && `Rombongan ${jamaah.rombongan}`, jamaah.regu && `Regu ${jamaah.regu}`].filter(Boolean).join(' · ') || '-'],
          ['Peran', [jamaah.ketua_rombongan && 'Ketua Rombongan', jamaah.ketua_regu && 'Ketua Regu'].filter(Boolean).join(', ') || '-'],
        ]
      : []),
    ...(jamaah.paket
      ? [
          ['Hotel Mekkah', [jamaah.paket.hotel_mekkah, jamaah.kamar_hotel_mekkah && `Kamar ${jamaah.kamar_hotel_mekkah}`].filter(Boolean).join(' · ') || '-'],
          ['Hotel Madinah', [jamaah.paket.hotel_madinah, jamaah.kamar_hotel_madinah && `Kamar ${jamaah.kamar_hotel_madinah}`].filter(Boolean).join(' · ') || '-'],
        ]
      : []),
    ['Keterangan', jamaah.keterangan || '-'],
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Detail Jamaah</h1>
        <div className="flex flex-wrap justify-end gap-2">
          <Link
            to={`/jamaah/new?dari=${id}`}
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700"
          >
            Daftarkan ke Paket Lain
          </Link>
          <Link
            to={`/jamaah/${id}/edit`}
            className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
          >
            Edit
          </Link>
          <Link
            to="/jamaah"
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
          {kelengkapanItems(jamaah.paket?.tipe).map((item) => (
            <CheckItem key={item.field} label={item.label} checked={jamaah[item.field]} />
          ))}
        </div>
      </div>

      {/* Same person (NIK) in other pakets */}
      {jamaah.paket_lain?.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-1">Paket Lain</h2>
          <p className="text-sm text-gray-500 mb-3">
            Jamaah dengan NIK yang sama juga terdaftar di paket berikut. Data diri dan dokumen identitas
            dipakai bersama; tagihan, manasik, dan bukti pembayaran dicatat terpisah per paket.
          </p>
          <ul className="divide-y divide-gray-100">
            {jamaah.paket_lain.map((p) => (
              <li key={p.jamaah_id} className="flex items-center justify-between gap-3 py-2">
                <Link to={`/jamaah/${p.jamaah_id}`} className="text-sm font-medium text-emerald-700 hover:underline">
                  {p.paket || 'Paket tidak ditemukan'}
                </Link>
                <StatusBadge status={p.status_pembayaran} />
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Tagihan */}
      {jamaah.tagihan && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Tagihan</h2>
          {jamaah.tagihan.harga_diatur ? (
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div className="rounded-lg border border-gray-200 p-4">
                <p className="text-sm text-gray-500">Harga</p>
                <p className="text-lg font-semibold text-gray-900">{formatRupiah(jamaah.tagihan.harga)}</p>
                <p className="text-xs text-gray-500 mt-1">
                  {jamaah.tagihan.pilihan_harga || 'Harga paket'}: {formatRupiah(jamaah.tagihan.harga_dasar)}
                </p>
                {jamaah.tagihan.potongan > 0 && (
                  <p className="text-xs text-gray-500">
                    Potongan −{formatRupiah(jamaah.tagihan.potongan)}
                    {jamaah.keterangan_potongan ? ` (${jamaah.keterangan_potongan})` : ''}
                  </p>
                )}
              </div>
              <div className="rounded-lg border border-gray-200 p-4">
                <p className="text-sm text-gray-500">Sudah dibayar</p>
                <p className="text-lg font-semibold text-gray-900">{formatRupiah(jamaah.tagihan.total_dibayar)}</p>
              </div>
              <div
                className={`rounded-lg border p-4 ${
                  jamaah.tagihan.sisa > 0 ? 'border-amber-200 bg-amber-50' : 'border-emerald-200 bg-emerald-50'
                }`}
              >
                <p className="text-sm text-gray-600">{jamaah.tagihan.sisa < 0 ? 'Kelebihan bayar' : 'Sisa tagihan'}</p>
                <p className="text-lg font-semibold text-gray-900">
                  {jamaah.tagihan.sisa > 0 ? formatRupiah(jamaah.tagihan.sisa) : jamaah.tagihan.sisa < 0 ? formatRupiah(-jamaah.tagihan.sisa) : 'Lunas'}
                </p>
              </div>
            </div>
          ) : (
            <p className="text-sm text-gray-600">
              Sudah dibayar <b>{formatRupiah(jamaah.tagihan.total_dibayar)}</b>. Harga paket belum diatur, jadi sisa
              tagihan belum bisa dihitung. Atur harga di menu{' '}
              <Link to="/paket" className="text-emerald-700 hover:underline">Paket</Link>.
            </p>
          )}
          {jamaah.tagihan.harga_diatur && jamaah.status_pembayaran === 'lunas' && jamaah.tagihan.sisa > 0 && (
            <p className="mt-3 text-sm text-amber-700">
              Status pembayaran Lunas, tetapi total bukti pembayaran masih kurang dari harga paket.
            </p>
          )}
          <p className="mt-3 text-xs text-gray-500">
            Dihitung dari nominal bukti DP dan bukti pelunasan yang diunggah di bawah.
          </p>
        </div>
      )}

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
            const docUrl = getDocumentUrl(id, doc.key, jamaah.dokumen_akses, cacheKey);
            return (
              <div
                key={doc.key}
                className="border border-gray-200 rounded-lg p-4"
              >
                <p className="text-sm font-medium text-gray-700 mb-2">
                  {doc.label}
                  {doc.shared && jamaah.paket_lain?.length > 0 && (
                    <span className="ml-2 text-xs font-normal text-gray-500">
                      dipakai bersama {jamaah.paket_lain.length + 1} paket
                    </span>
                  )}
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
                      href={getDocumentDownloadUrl(id, doc.key, jamaah.dokumen_akses)}
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
                        setNominalUpload({ docKey: doc.key, label: doc.label, file });
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
                    const fileUrl = getMultiDocumentUrl(id, doc.key, idx, jamaah.dokumen_akses, cacheKey);
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
                          href={getMultiDocumentDownloadUrl(id, doc.key, idx, jamaah.dokumen_akses)}
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

      {nominalUpload && (
        <NominalDialog
          title={`${nominalUpload.label} — nominal pembayaran`}
          file={nominalUpload.file}
          sisa={jamaah.tagihan?.harga_diatur ? jamaah.tagihan.sisa : 0}
          onCancel={() => setNominalUpload(null)}
          onSubmit={async (nominal) => {
            await handleUpload(nominalUpload.docKey, nominalUpload.file, nominal);
            setNominalUpload(null);
          }}
        />
      )}

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
