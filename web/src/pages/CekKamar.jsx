import { Link } from 'react-router-dom';
import CekKamarCard from '../components/CekKamarCard';

export default function CekKamar() {
  return (
    <div className="min-h-screen bg-gray-50 px-4 py-10">
      <div className="max-w-md mx-auto">
        <h1 className="text-2xl font-bold text-center text-emerald-700">Syairullah Haji & Umroh</h1>
        <p className="text-center text-gray-500 mt-1 mb-6">Cek Kamar Manasik</p>

        <CekKamarCard />

        <p className="mt-6 text-center text-sm text-gray-500">
          <Link to="/" className="hover:underline">
            ← Kembali ke beranda
          </Link>
        </p>
      </div>
    </div>
  );
}
