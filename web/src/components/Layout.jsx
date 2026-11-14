import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/auth';

const navLinks = [
  { to: '/dashboard', label: 'Dashboard' },
  { to: '/jamaah', label: 'Jamaah' },
  { to: '/manasik', label: 'Manasik' },
  { to: '/tagihan', label: 'Tagihan' },
  { to: '/paket', label: 'Paket' },
  { to: '/pengumuman', label: 'Pengumuman' },
  { to: '/riwayat', label: 'Riwayat' },
  { to: '/admin', label: 'Admin' },
];

export default function Layout({ children }) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const isActive = (path) => location.pathname.startsWith(path);

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-emerald-700 text-white shadow-lg print:hidden">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between lg:h-16 py-2 lg:py-0 gap-1">
            <div className="flex items-center justify-between">
              <Link to="/dashboard" className="text-lg lg:text-xl font-bold tracking-wide">
                Syairullah Haji & Umroh
              </Link>
              <button
                onClick={() => { logout(); navigate('/login'); }}
                className="lg:hidden px-3 py-1.5 rounded-md text-sm font-medium text-emerald-100 hover:bg-emerald-600 hover:text-white"
              >
                Logout
              </button>
            </div>
            <div className="flex items-center min-w-0">
              <div className="flex items-center space-x-1 overflow-x-auto -mx-1 px-1 pb-1 lg:pb-0">
                {navLinks.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    className={`shrink-0 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                      isActive(link.to)
                        ? 'bg-emerald-800 text-white'
                        : 'text-emerald-100 hover:bg-emerald-600 hover:text-white'
                    }`}
                  >
                    {link.label}
                  </Link>
                ))}
              </div>
              <div className="hidden lg:flex ml-6 pl-6 border-l border-emerald-500 items-center space-x-3">
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
      {user?.password_default && (
        <div role="alert" className="bg-red-600 text-white print:hidden">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-2 text-sm flex flex-wrap items-center justify-between gap-2">
            <span>
              <b>Password Anda masih default (admin123).</b> Siapa pun yang tahu password ini bisa masuk sebagai admin.
            </span>
            {user.role === 'super' && (
              <Link to="/admin" className="font-semibold underline whitespace-nowrap">
                Ganti password sekarang →
              </Link>
            )}
          </div>
        </div>
      )}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 lg:py-8 print:p-0 print:max-w-none">
        {children}
      </main>
    </div>
  );
}
