import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchPublicPengumuman } from '../api';
import { useAuth } from '../contexts/AuthContext';
import CekKamarCard from '../components/CekKamarCard';

function formatTanggal(dateStr) {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function PengumumanItem({ item }) {
  return (
    <article
      className={`rounded-lg border p-5 ${
        item.disematkan ? 'bg-amber-50 border-amber-200' : 'bg-white border-gray-200'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="font-semibold text-gray-900">{item.judul}</h3>
        {item.disematkan && (
          <span className="shrink-0 text-xs font-medium text-amber-800 bg-amber-100 rounded-full px-2 py-0.5">
            Penting
          </span>
        )}
      </div>
      <p className="text-xs text-gray-500 mt-1">{formatTanggal(item.created_at)}</p>
      <p className="mt-3 text-sm text-gray-700 whitespace-pre-line break-words">{item.isi}</p>
    </article>
  );
}

export default function Home() {
  const { user } = useAuth();
  const [pengumuman, setPengumuman] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchPublicPengumuman()
      .then(setPengumuman)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-emerald-700 text-white">
        <div className="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between">
          <span className="text-lg sm:text-xl font-bold tracking-wide">Syairullah Haji & Umroh</span>
          {/* Admin entry is only shown to logged-in admins; others reach /login directly. */}
          {user && (
            <Link to="/dashboard" className="text-sm font-medium text-emerald-100 hover:text-white">
              Dashboard Admin
            </Link>
          )}
        </div>
      </header>

      <section className="bg-emerald-700 text-white pb-10">
        <div className="max-w-5xl mx-auto px-4 pt-6">
          <h1 className="text-2xl sm:text-3xl font-bold">Assalamu'alaikum, Jamaah</h1>
          <p className="mt-2 text-emerald-100 max-w-2xl">
            Cek gedung dan kamar manasik Anda di asrama haji, dan ikuti pengumuman terbaru dari
            Syairullah Haji & Umroh.
          </p>
        </div>
      </section>

      <main className="max-w-5xl mx-auto px-4 -mt-6 pb-12 grid gap-8 md:grid-cols-5">
        <section className="md:col-span-2" aria-labelledby="cek-kamar-title">
          <h2 id="cek-kamar-title" className="sr-only">
            Cek Kamar Manasik
          </h2>
          <div className="md:sticky md:top-6">
            <CekKamarCard />
          </div>
        </section>

        <section className="md:col-span-3" aria-labelledby="pengumuman-title">
          <h2 id="pengumuman-title" className="text-lg font-semibold text-gray-900 mb-4 md:mt-8">
            Pengumuman
          </h2>
          {loading ? (
            <p className="text-sm text-gray-500">Memuat pengumuman...</p>
          ) : error ? (
            <p className="text-sm text-red-600">{error}</p>
          ) : pengumuman.length === 0 ? (
            <p className="text-sm text-gray-500">Belum ada pengumuman.</p>
          ) : (
            <div className="space-y-4">
              {pengumuman.map((item) => (
                <PengumumanItem key={item.id} item={item} />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
