import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import QrScanner from 'qr-scanner';
import { fetchSesiDetail, markHadir, unmarkHadir } from '../api';

const collator = new Intl.Collator('id', { numeric: true, sensitivity: 'base' });

function formatJam(dateStr) {
  return new Date(dateStr).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
}

function penempatan(j) {
  return [
    j.gedung_manasik && j.kamar_manasik ? `${j.gedung_manasik} / ${j.kamar_manasik}` : j.gedung_manasik || j.kamar_manasik,
    j.rombongan && `R${j.rombongan}`,
    j.regu && `Regu ${j.regu}`,
  ]
    .filter(Boolean)
    .join(' · ');
}

// Short audible + haptic feedback so petugas don't need to look at the screen.
function feedback(ok) {
  try {
    navigator.vibrate?.(ok ? 80 : [60, 60, 60]);
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.frequency.value = ok ? 880 : 220;
    gain.gain.value = 0.15;
    osc.connect(gain).connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + (ok ? 0.12 : 0.3));
    osc.onended = () => ctx.close();
  } catch {
    // feedback is best-effort
  }
}

function Scanner({ onDecode }) {
  const videoRef = useRef(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return undefined;
    const scanner = new QrScanner(video, (result) => onDecode(result.data), {
      returnDetailedScanResult: true,
      preferredCamera: 'environment',
      highlightScanRegion: true,
      highlightCodeOutline: true,
      maxScansPerSecond: 5,
    });
    scanner.start().catch((err) => {
      const msg = String(err?.message || err || '');
      if (!window.isSecureContext) {
        setError('Kamera hanya bisa dipakai lewat HTTPS (atau localhost). Gunakan pencarian manual di bawah.');
      } else if (/permission|notallowed|denied/i.test(msg)) {
        setError('Izin kamera ditolak. Izinkan akses kamera di pengaturan browser, lalu muat ulang halaman.');
      } else {
        setError('Kamera tidak dapat dibuka. Gunakan pencarian manual di bawah.');
      }
    });
    return () => {
      scanner.stop();
      scanner.destroy();
    };
  }, [onDecode]);

  return (
    <div>
      <div className="relative mx-auto w-full max-w-sm aspect-square overflow-hidden rounded-lg bg-black">
        <video ref={videoRef} className="w-full h-full object-cover" muted playsInline />
      </div>
      {error && <div className="mt-3 p-3 bg-amber-50 text-amber-800 rounded-md text-sm">{error}</div>}
    </div>
  );
}

export default function SesiManasik() {
  const { id } = useParams();
  const [data, setData] = useState(null);
  const [error, setError] = useState('');
  const [scanning, setScanning] = useState(false);
  const [last, setLast] = useState(null); // {ok, status, title, detail}
  const [tab, setTab] = useState('belum');
  const [search, setSearch] = useState('');
  const [busyId, setBusyId] = useState('');
  const busyRef = useRef(false);
  const recentRef = useRef(new Map()); // token -> timestamp, to ignore repeated frames

  const load = useCallback(() => {
    return fetchSesiDetail(id)
      .then((d) => {
        setData(d);
        setError('');
      })
      .catch((err) => setError(err.message));
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  const applyHadir = (jamaah) => {
    setData((prev) =>
      prev && {
        ...prev,
        jamaah: prev.jamaah.map((j) => (j.id === jamaah.id ? { ...j, hadir: jamaah.hadir } : j)),
      },
    );
  };

  const record = useCallback(async (payload) => {
    try {
      const res = await markHadir(id, payload);
      const j = res.jamaah;
      applyHadir(j);
      const isNew = res.status === 'baru';
      feedback(true);
      setLast({
        ok: true,
        status: isNew ? 'baru' : 'sudah',
        title: j.nama,
        detail: isNew
          ? penempatan(j) || 'Hadir tercatat'
          : `Sudah tercatat pukul ${formatJam(j.hadir.waktu)}${j.hadir.dicatat_oleh ? ` oleh ${j.hadir.dicatat_oleh}` : ''}`,
      });
    } catch (err) {
      feedback(false);
      setLast({ ok: false, title: 'Gagal', detail: err.message });
    }
  }, [id]);

  const onDecode = useCallback(
    async (token) => {
      const now = Date.now();
      const seen = recentRef.current.get(token);
      if (busyRef.current || (seen && now - seen < 4000)) return;
      recentRef.current.set(token, now);
      busyRef.current = true;
      await record({ token });
      busyRef.current = false;
    },
    [record],
  );

  const manual = async (j) => {
    setBusyId(j.id);
    await record({ jamaah_id: j.id });
    setBusyId('');
  };

  const undo = async (j) => {
    if (!confirm(`Batalkan kehadiran ${j.nama}?`)) return;
    setBusyId(j.id);
    try {
      await unmarkHadir(id, j.id);
      applyHadir({ ...j, hadir: null });
    } catch (err) {
      setError(err.message);
    } finally {
      setBusyId('');
    }
  };

  const { hadir, belum } = useMemo(() => {
    const list = data?.jamaah || [];
    const q = search.trim().toLowerCase();
    const match = (j) => !q || j.nama.toLowerCase().includes(q) || j.nik.includes(q);
    return {
      hadir: list
        .filter((j) => j.hadir && match(j))
        .sort((a, b) => new Date(b.hadir.waktu) - new Date(a.hadir.waktu)),
      belum: list.filter((j) => !j.hadir && match(j)).sort((a, b) => collator.compare(a.nama, b.nama)),
    };
  }, [data, search]);

  if (error && !data) return <p className="text-red-600">{error}</p>;
  if (!data) return <p className="text-gray-500">Memuat data...</p>;

  const total = data.jamaah.length;
  const jumlahHadir = data.jamaah.filter((j) => j.hadir).length;
  const pct = total ? Math.round((jumlahHadir / total) * 100) : 0;
  const shown = tab === 'hadir' ? hadir : belum;

  return (
    <div className="max-w-3xl mx-auto">
      <Link
        to={`/manasik?tab=absensi&paket=${data.sesi.paket_id}`}
        className="text-sm text-emerald-700 hover:underline"
      >
        ← Kembali ke daftar sesi
      </Link>

      <div className="mt-2 mb-4">
        <h1 className="text-2xl font-bold text-gray-900">{data.sesi.nama}</h1>
        <p className="text-sm text-gray-500">
          {data.paket?.label} ·{' '}
          {new Date(data.sesi.tanggal).toLocaleDateString('id-ID', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
            year: 'numeric',
            timeZone: 'UTC',
          })}
        </p>
      </div>

      <div className="bg-white shadow rounded-lg p-4 mb-4">
        <div className="flex justify-between items-baseline">
          <span className="text-sm text-gray-600">Kehadiran</span>
          <span className="text-2xl font-bold text-gray-900">
            {jumlahHadir} <span className="text-base font-normal text-gray-500">/ {total} ({pct}%)</span>
          </span>
        </div>
        <div className="mt-2 h-2.5 rounded-full bg-gray-100 overflow-hidden">
          <div className="h-full bg-emerald-500 transition-all" style={{ width: `${pct}%` }} />
        </div>
      </div>

      <div className="bg-white shadow rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between gap-3">
          <h2 className="font-semibold text-gray-900">Scan QR Jamaah</h2>
          <button
            onClick={() => setScanning((s) => !s)}
            className={`px-4 py-2 text-sm font-medium rounded-md ${
              scanning ? 'bg-gray-200 text-gray-800 hover:bg-gray-300' : 'bg-emerald-600 text-white hover:bg-emerald-700'
            }`}
          >
            {scanning ? 'Tutup Kamera' : 'Buka Kamera'}
          </button>
        </div>
        {scanning && (
          <div className="mt-4">
            <Scanner onDecode={onDecode} />
          </div>
        )}
        {last && (
          <div
            role="status"
            aria-live="polite"
            className={`mt-4 rounded-lg p-4 border ${
              !last.ok
                ? 'bg-red-50 border-red-200 text-red-800'
                : last.status === 'baru'
                  ? 'bg-emerald-50 border-emerald-200 text-emerald-900'
                  : 'bg-amber-50 border-amber-200 text-amber-900'
            }`}
          >
            <p className="text-xs font-semibold uppercase tracking-wide">
              {!last.ok ? 'Tidak tercatat' : last.status === 'baru' ? '✓ Hadir tercatat' : 'Sudah hadir sebelumnya'}
            </p>
            <p className="text-lg font-semibold mt-0.5">{last.title}</p>
            <p className="text-sm">{last.detail}</p>
          </div>
        )}
      </div>

      {error && <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}

      <div className="bg-white shadow rounded-lg">
        <div className="p-4 border-b border-gray-100 space-y-3">
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari nama atau NIK (untuk absen manual)"
            aria-label="Cari jamaah"
            className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
          />
          <div className="inline-flex rounded-md border border-gray-300 overflow-hidden" role="group">
            {[
              ['belum', `Belum hadir (${belum.length})`],
              ['hadir', `Hadir (${hadir.length})`],
            ].map(([key, label]) => (
              <button
                key={key}
                onClick={() => setTab(key)}
                aria-pressed={tab === key}
                className={`px-3 py-1.5 text-sm font-medium ${
                  tab === key ? 'bg-emerald-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
                }`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
        {shown.length === 0 ? (
          <p className="p-4 text-sm text-gray-500">
            {search ? 'Tidak ada jamaah yang cocok.' : tab === 'belum' ? 'Semua jamaah sudah hadir.' : 'Belum ada yang hadir.'}
          </p>
        ) : (
          <ul className="divide-y divide-gray-100">
            {shown.map((j) => (
              <li key={j.id} className="px-4 py-3 flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{j.nama}</p>
                  <p className="text-xs text-gray-500 truncate">
                    {tab === 'hadir'
                      ? `${formatJam(j.hadir.waktu)} · ${j.hadir.metode === 'qr' ? 'QR' : 'manual'}${
                          j.hadir.dicatat_oleh ? ` · ${j.hadir.dicatat_oleh}` : ''
                        }`
                      : penempatan(j) || j.nik}
                  </p>
                </div>
                {tab === 'belum' ? (
                  <button
                    onClick={() => manual(j)}
                    disabled={busyId === j.id}
                    className="shrink-0 px-3 py-1.5 text-sm font-medium rounded-md bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-50"
                  >
                    Hadir
                  </button>
                ) : (
                  <button
                    onClick={() => undo(j)}
                    disabled={busyId === j.id}
                    className="shrink-0 px-3 py-1.5 text-sm font-medium rounded-md text-red-600 hover:bg-red-50 disabled:opacity-50"
                  >
                    Batalkan
                  </button>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
