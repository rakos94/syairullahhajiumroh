import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  fetchPaketList,
  fetchManasikJamaah,
  downloadManasikExcel,
  importManasik,
  fetchSesiManasik,
  createSesiManasik,
  deleteSesiManasik,
} from '../api';

const collator = new Intl.Collator('id', { numeric: true, sensitivity: 'base' });
const cmp = (a, b) => {
  if (!a && b) return 1; // empty last
  if (a && !b) return -1;
  return collator.compare(a || '', b || '');
};

const btn =
  'inline-flex items-center justify-center px-3 py-2 text-sm font-medium rounded-md disabled:opacity-50';
const btnPrimary = `${btn} bg-emerald-600 text-white hover:bg-emerald-700`;
const btnSecondary = `${btn} bg-white border border-gray-300 text-gray-700 hover:bg-gray-50`;
const inputClass =
  'px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500';

function KetuaBadges({ j }) {
  return (
    <>
      {j.ketua_rombongan && (
        <span className="ml-2 text-[11px] font-medium text-purple-800 bg-purple-100 rounded-full px-1.5 py-0.5 print:border print:border-gray-400 print:bg-transparent">
          Ketua Rombongan
        </span>
      )}
      {j.ketua_regu && (
        <span className="ml-2 text-[11px] font-medium text-blue-800 bg-blue-100 rounded-full px-1.5 py-0.5 print:border print:border-gray-400 print:bg-transparent">
          Ketua Regu
        </span>
      )}
    </>
  );
}

// Groups jamaah into [{title, subgroups: [{title, items}]}].
function groupJamaah(list, view) {
  const outerKey = view === 'kamar' ? 'gedung_manasik' : 'rombongan';
  const innerKey = view === 'kamar' ? 'kamar_manasik' : 'regu';
  const outerLabel = (v) => (view === 'kamar' ? v || 'Belum ada gedung' : v ? `Rombongan ${v}` : 'Belum ada rombongan');
  const innerLabel = (v) => (view === 'kamar' ? (v ? `Kamar ${v}` : 'Belum ada kamar') : v ? `Regu ${v}` : 'Belum ada regu');

  const outer = new Map();
  for (const j of list) {
    const o = j[outerKey] || '';
    const i = j[innerKey] || '';
    if (!outer.has(o)) outer.set(o, new Map());
    const inner = outer.get(o);
    if (!inner.has(i)) inner.set(i, []);
    inner.get(i).push(j);
  }
  return [...outer.entries()]
    .sort(([a], [b]) => cmp(a, b))
    .map(([o, inner]) => ({
      key: o,
      title: outerLabel(o),
      subgroups: [...inner.entries()]
        .sort(([a], [b]) => cmp(a, b))
        .map(([i, items]) => ({
          key: i,
          title: innerLabel(i),
          items: [...items].sort((a, b) => collator.compare(a.nama, b.nama)),
        })),
    }));
}

function ImportPanel({ paketId, onDone, onClose }) {
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [doneMsg, setDoneMsg] = useState('');

  const runPreview = async (f) => {
    setFile(f);
    setPreview(null);
    setDoneMsg('');
    setError('');
    if (!f) return;
    setBusy(true);
    try {
      setPreview(await importManasik(paketId, f, true));
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const apply = async () => {
    setBusy(true);
    setError('');
    try {
      const res = await importManasik(paketId, file, false);
      setDoneMsg(`${res.summary.ubah} jamaah berhasil diperbarui.`);
      setPreview(null);
      setFile(null);
      onDone();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const shown = preview ? preview.rows.filter((r) => r.status !== 'sama') : [];

  return (
    <div className="bg-white shadow rounded-lg p-5 mb-6 print:hidden">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-semibold text-gray-900">Import Penempatan dari Excel</h3>
          <p className="text-sm text-gray-500 mt-1">
            Paling mudah: klik <b>Unduh Excel</b>, isi kolom Rombongan, Regu, Gedung, dan Kamar, lalu
            unggah kembali di sini. Kolom NIK wajib ada. Sel yang dikosongkan akan menghapus nilainya;
            kolom yang tidak ada di file tidak diubah. Format .xlsx atau .csv.
          </p>
        </div>
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600" aria-label="Tutup import">
          ✕
        </button>
      </div>

      <input
        type="file"
        accept=".xlsx,.csv"
        onChange={(e) => runPreview(e.target.files?.[0] || null)}
        className="mt-4 block text-sm text-gray-700 file:mr-3 file:px-3 file:py-2 file:rounded-md file:border-0 file:bg-emerald-50 file:text-emerald-800 file:font-medium hover:file:bg-emerald-100"
      />

      {busy && <p className="mt-3 text-sm text-gray-500">Memproses...</p>}
      {error && <div className="mt-3 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}
      {doneMsg && <div className="mt-3 p-3 bg-emerald-50 text-emerald-800 rounded-md text-sm">{doneMsg}</div>}

      {preview && (
        <div className="mt-4">
          <div className="flex flex-wrap gap-2 text-sm">
            <span className="px-2 py-1 rounded-md bg-emerald-50 text-emerald-800">
              {preview.summary.ubah} akan diubah
            </span>
            <span className="px-2 py-1 rounded-md bg-gray-100 text-gray-700">
              {preview.summary.sama} tidak berubah
            </span>
            {preview.summary.error > 0 && (
              <span className="px-2 py-1 rounded-md bg-red-50 text-red-700">
                {preview.summary.error} baris error (akan dilewati)
              </span>
            )}
          </div>

          {shown.length > 0 && (
            <div className="mt-3 max-h-80 overflow-auto border border-gray-200 rounded-md">
              <table className="min-w-full text-sm">
                <thead className="bg-gray-50 sticky top-0">
                  <tr className="text-left text-xs uppercase text-gray-500">
                    <th className="px-3 py-2">Baris</th>
                    <th className="px-3 py-2">Jamaah</th>
                    <th className="px-3 py-2">Perubahan / Error</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {shown.map((r) => (
                    <tr key={r.row} className={r.status === 'error' ? 'bg-red-50/50' : ''}>
                      <td className="px-3 py-2 text-gray-500">{r.row}</td>
                      <td className="px-3 py-2">
                        <div className="font-medium text-gray-900">{r.nama || '-'}</div>
                        <div className="text-xs text-gray-500">{r.nik}</div>
                      </td>
                      <td className="px-3 py-2">
                        {r.status === 'error' ? (
                          <span className="text-red-700">{r.error}</span>
                        ) : (
                          <ul className="space-y-0.5">
                            {r.changes.map((c) => (
                              <li key={c.field}>
                                <span className="text-gray-500">{c.field}:</span>{' '}
                                <span className="line-through text-gray-400">{c.old_value || 'kosong'}</span>{' '}
                                → <span className="font-medium text-gray-900">{c.new_value || 'kosong'}</span>
                              </li>
                            ))}
                          </ul>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <div className="mt-4 flex gap-2">
            <button onClick={apply} disabled={busy || preview.summary.ubah === 0} className={btnPrimary}>
              Terapkan {preview.summary.ubah} perubahan
            </button>
            <button onClick={() => runPreview(null)} disabled={busy} className={btnSecondary}>
              Batal
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function KamarTab({ paket }) {
  const [list, setList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [view, setView] = useState('kamar');
  const [showImport, setShowImport] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    fetchManasikJamaah(paket.id)
      .then((res) => {
        if (cancelled) return;
        setList(res.jamaah || []);
        setError('');
      })
      .catch((err) => !cancelled && setError(err.message))
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [paket.id, reloadKey]);

  const groups = useMemo(() => groupJamaah(list, view), [list, view]);
  const withKamar = list.filter((j) => j.gedung_manasik && j.kamar_manasik).length;
  const withRegu = list.filter((j) => j.rombongan && j.regu).length;

  const download = async () => {
    setDownloading(true);
    try {
      await downloadManasikExcel(paket.id);
    } catch (err) {
      setError(err.message);
    } finally {
      setDownloading(false);
    }
  };

  if (loading) return <p className="text-gray-500">Memuat data...</p>;

  return (
    <div>
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 mb-4 print:hidden">
        <div className="flex flex-wrap gap-2 text-sm">
          <span className="px-2.5 py-1 rounded-md bg-white border border-gray-200">{list.length} jamaah</span>
          <span className="px-2.5 py-1 rounded-md bg-white border border-gray-200">
            {withKamar} sudah dapat kamar
          </span>
          <span className="px-2.5 py-1 rounded-md bg-white border border-gray-200">{withRegu} sudah dapat regu</span>
          {list.length - withKamar > 0 && (
            <Link
              to={`/jamaah?paket_id=${paket.id}&kelengkapan=manasik_belum`}
              className="px-2.5 py-1 rounded-md bg-amber-50 border border-amber-200 text-amber-800 hover:bg-amber-100"
            >
              {list.length - withKamar} belum dapat kamar →
            </Link>
          )}
        </div>
        <div className="flex flex-wrap gap-2">
          <div className="inline-flex rounded-md border border-gray-300 overflow-hidden" role="group" aria-label="Kelompokkan">
            {[
              ['kamar', 'Per Kamar'],
              ['regu', 'Per Regu'],
            ].map(([v, label]) => (
              <button
                key={v}
                onClick={() => setView(v)}
                aria-pressed={view === v}
                className={`px-3 py-2 text-sm font-medium ${
                  view === v ? 'bg-emerald-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
                }`}
              >
                {label}
              </button>
            ))}
          </div>
          <button onClick={() => window.print()} className={btnSecondary}>
            Cetak
          </button>
          <button onClick={download} disabled={downloading} className={btnSecondary}>
            {downloading ? 'Mengunduh...' : 'Unduh Excel'}
          </button>
          <button onClick={() => setShowImport((v) => !v)} className={btnPrimary}>
            Import Excel
          </button>
        </div>
      </div>

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      {showImport && (
        <ImportPanel
          paketId={paket.id}
          onDone={() => setReloadKey((k) => k + 1)}
          onClose={() => setShowImport(false)}
        />
      )}

      <div className="hidden print:block mb-4">
        <h1 className="text-xl font-bold">
          Daftar {view === 'kamar' ? 'Kamar' : 'Regu'} Manasik — {paket.label}
        </h1>
        <p className="text-sm">Dicetak {new Date().toLocaleString('id-ID')}</p>
      </div>

      {list.length === 0 ? (
        <p className="text-gray-500">Belum ada jamaah di paket ini.</p>
      ) : (
        <div className="space-y-6 print:space-y-4">
          {groups.map((g) => (
            <section key={g.key} className="break-inside-avoid-page">
              <h2 className="text-lg font-semibold text-gray-900 mb-2">
                {g.title}{' '}
                <span className="text-sm font-normal text-gray-500">
                  ({g.subgroups.reduce((n, s) => n + s.items.length, 0)} jamaah)
                </span>
              </h2>
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3 print:grid-cols-2">
                {g.subgroups.map((sg) => (
                  <div
                    key={sg.key}
                    className={`bg-white rounded-lg shadow print:shadow-none print:border print:border-gray-400 break-inside-avoid ${
                      sg.key ? '' : 'ring-1 ring-amber-200'
                    }`}
                  >
                    <div className="px-4 py-2 border-b border-gray-100 flex justify-between items-center">
                      <span className={`font-medium ${sg.key ? 'text-gray-900' : 'text-amber-700'}`}>{sg.title}</span>
                      <span className="text-xs text-gray-500">{sg.items.length} orang</span>
                    </div>
                    <ol className="px-4 py-2 text-sm space-y-1">
                      {sg.items.map((j, idx) => (
                        <li key={j.id} className="flex justify-between gap-2">
                          <span>
                            <span className="text-gray-400 mr-1">{idx + 1}.</span>
                            <Link to={`/jamaah/${j.id}`} className="text-gray-900 hover:underline print:no-underline">
                              {j.nama}
                            </Link>
                            <span className="ml-1 text-xs text-gray-500">({j.jenis_kelamin === 'perempuan' ? 'P' : 'L'})</span>
                            <KetuaBadges j={j} />
                          </span>
                          <span className="text-xs text-gray-500 whitespace-nowrap">
                            {view === 'kamar'
                              ? [j.rombongan && `R${j.rombongan}`, j.regu && `Regu ${j.regu}`].filter(Boolean).join(' · ')
                              : [j.gedung_manasik, j.kamar_manasik].filter(Boolean).join(' / ')}
                          </span>
                        </li>
                      ))}
                    </ol>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}

function AbsensiTab({ paket }) {
  const [sesi, setSesi] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [nama, setNama] = useState('');
  const [tanggal, setTanggal] = useState(() => new Date().toISOString().slice(0, 10));
  const [saving, setSaving] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    Promise.all([fetchSesiManasik(paket.id), fetchManasikJamaah(paket.id)])
      .then(([s, j]) => {
        if (cancelled) return;
        setSesi(s);
        setTotal((j.jamaah || []).length);
        setError('');
      })
      .catch((err) => !cancelled && setError(err.message))
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [paket.id, reloadKey]);

  const create = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      await createSesiManasik({ paket_id: paket.id, nama, tanggal });
      setNama('');
      setReloadKey((k) => k + 1);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const remove = async (s) => {
    if (!confirm(`Hapus sesi "${s.nama}" beserta ${s.jumlah_hadir} data kehadirannya?`)) return;
    try {
      await deleteSesiManasik(s.id);
      setReloadKey((k) => k + 1);
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading) return <p className="text-gray-500">Memuat data...</p>;

  return (
    <div>
      <form onSubmit={create} className="bg-white shadow rounded-lg p-4 mb-6 flex flex-col sm:flex-row gap-3 sm:items-end">
        <div className="flex-1">
          <label htmlFor="sesi-nama" className="block text-sm font-medium text-gray-700 mb-1">
            Nama sesi baru
          </label>
          <input
            id="sesi-nama"
            value={nama}
            onChange={(e) => setNama(e.target.value)}
            placeholder="Contoh: Hari 1 - Pagi"
            maxLength={100}
            required
            className={`${inputClass} w-full`}
          />
        </div>
        <div>
          <label htmlFor="sesi-tanggal" className="block text-sm font-medium text-gray-700 mb-1">
            Tanggal
          </label>
          <input
            id="sesi-tanggal"
            type="date"
            value={tanggal}
            onChange={(e) => setTanggal(e.target.value)}
            required
            className={`${inputClass} w-full`}
          />
        </div>
        <button type="submit" disabled={saving} className={btnPrimary}>
          {saving ? 'Menyimpan...' : '+ Tambah Sesi'}
        </button>
      </form>

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      {sesi.length === 0 ? (
        <p className="text-gray-500">Belum ada sesi manasik. Tambahkan sesi untuk mulai mencatat kehadiran.</p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {sesi.map((s) => {
            const pct = total ? Math.round((s.jumlah_hadir / total) * 100) : 0;
            return (
              <div key={s.id} className="bg-white shadow rounded-lg p-4">
                <div className="flex justify-between gap-2">
                  <div>
                    <p className="font-semibold text-gray-900">{s.nama}</p>
                    <p className="text-xs text-gray-500">
                      {new Date(s.tanggal).toLocaleDateString('id-ID', {
                        weekday: 'long',
                        day: 'numeric',
                        month: 'long',
                        year: 'numeric',
                        timeZone: 'UTC',
                      })}
                    </p>
                  </div>
                  <button onClick={() => remove(s)} className="text-xs text-red-600 hover:text-red-800 self-start">
                    Hapus
                  </button>
                </div>
                <div className="mt-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Hadir</span>
                    <span className="font-medium text-gray-900">
                      {s.jumlah_hadir} / {total}
                    </span>
                  </div>
                  <div className="mt-1 h-2 rounded-full bg-gray-100 overflow-hidden">
                    <div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} />
                  </div>
                </div>
                <Link to={`/manasik/sesi/${s.id}`} className={`${btnPrimary} w-full mt-4`}>
                  Buka Absensi & Scan QR
                </Link>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default function Manasik() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [pakets, setPakets] = useState(null);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchPaketList('haji')
      .then((list) => setPakets([...list].sort((a, b) => b.tahun - a.tahun)))
      .catch((err) => setError(err.message));
  }, []);

  const tab = searchParams.get('tab') === 'absensi' ? 'absensi' : 'kamar';
  const paketId = searchParams.get('paket') || pakets?.[0]?.id || '';
  const paket = pakets?.find((p) => p.id === paketId);

  const setParam = (key, value) => {
    const next = new URLSearchParams(searchParams);
    next.set(key, value);
    setSearchParams(next, { replace: true });
  };

  return (
    <div>
      <div className="sm:flex sm:items-center sm:justify-between gap-4 mb-4 print:hidden">
        <h1 className="text-2xl font-bold text-gray-900">Manasik</h1>
        {pakets && pakets.length > 0 && (
          <select
            value={paketId}
            onChange={(e) => setParam('paket', e.target.value)}
            aria-label="Pilih paket haji"
            className={`${inputClass} mt-3 sm:mt-0`}
          >
            {pakets.map((p) => (
              <option key={p.id} value={p.id}>
                {p.label}
              </option>
            ))}
          </select>
        )}
      </div>

      <div className="border-b border-gray-200 mb-6 print:hidden">
        <nav className="-mb-px flex gap-6" aria-label="Tab manasik">
          {[
            ['kamar', 'Daftar Kamar & Regu'],
            ['absensi', 'Absensi'],
          ].map(([key, label]) => (
            <button
              key={key}
              onClick={() => setParam('tab', key)}
              aria-current={tab === key ? 'page' : undefined}
              className={`py-2 text-sm font-medium border-b-2 ${
                tab === key
                  ? 'border-emerald-600 text-emerald-700'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {label}
            </button>
          ))}
        </nav>
      </div>

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      {!pakets ? (
        <p className="text-gray-500">Memuat data...</p>
      ) : pakets.length === 0 ? (
        <p className="text-gray-500">
          Belum ada paket haji. <Link to="/paket" className="text-emerald-700 hover:underline">Tambah paket</Link>{' '}
          terlebih dahulu.
        </p>
      ) : !paket ? (
        <p className="text-gray-500">Paket tidak ditemukan.</p>
      ) : tab === 'kamar' ? (
        <KamarTab key={paket.id} paket={paket} />
      ) : (
        <AbsensiTab key={paket.id} paket={paket} />
      )}
    </div>
  );
}
