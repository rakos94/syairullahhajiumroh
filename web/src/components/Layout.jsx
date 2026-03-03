import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const navLinks = [
  { to: '/', label: 'Jamaah' },
  { to: '/paket', label: 'Paket' },
  { to: '/admin', label: 'Admin' },
];

export default function Layout({ children }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const isActive = (path) =>
    path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-emerald-700 text-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link to="/" className="text-xl font-bold tracking-wide">
              Syairullah Haji & Umroh
            </Link>
            <div className="flex items-center">
              <div className="flex items-center space-x-1">
                {navLinks.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                      isActive(link.to)
                        ? 'bg-emerald-800 text-white'
                        : 'text-emerald-100 hover:bg-emerald-600 hover:text-white'
                    }`}
                  >
                    {link.label}
                  </Link>
                ))}
              </div>
              <div className="ml-6 pl-6 border-l border-emerald-500 flex items-center space-x-3">
                <span className="text-sm text-emerald-200">{user?.username}</span>
                <button
                  onClick={() => { logout(); navigate('/login'); }}
                  className="px-3 py-1.5 rounded-md text-sm font-medium text-emerald-100 hover:bg-emerald-600 hover:text-white transition-colors"
                >
                  Logout
                </button>
              </div>
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
