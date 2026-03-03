import { Link, useLocation } from 'react-router-dom';

export default function Layout({ children }) {
  const location = useLocation();

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-emerald-700 text-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link to="/" className="text-xl font-bold tracking-wide">
              Syairullah Haji & Umroh
            </Link>
            <div className="flex space-x-4">
              <Link
                to="/"
                className={`px-3 py-2 rounded-md text-sm font-medium ${
                  location.pathname === '/'
                    ? 'bg-emerald-800'
                    : 'hover:bg-emerald-600'
                }`}
              >
                Jamaah
              </Link>
              <Link
                to="/paket"
                className={`px-3 py-2 rounded-md text-sm font-medium ${
                  location.pathname === '/paket'
                    ? 'bg-emerald-800'
                    : 'hover:bg-emerald-600'
                }`}
              >
                Paket
              </Link>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
}
