import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import JamaahList from './pages/JamaahList';
import JamaahForm from './pages/JamaahForm';
import JamaahDetail from './pages/JamaahDetail';
import PaketList from './pages/PaketList';
import AdminList from './pages/AdminList';
import AuditLog from './pages/AuditLog';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import CekKamar from './pages/CekKamar';
import Home from './pages/Home';
import PengumumanList from './pages/PengumumanList';
import Manasik from './pages/Manasik';
import SesiManasik from './pages/SesiManasik';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Home />} />
      <Route path="/cek-kamar" element={<CekKamar />} />
      <Route
        path="/*"
        element={
          <ProtectedRoute>
            <Layout>
              <Routes>
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/jamaah" element={<JamaahList />} />
                <Route path="/jamaah/new" element={<JamaahForm />} />
                <Route path="/jamaah/:id/edit" element={<JamaahForm />} />
                <Route path="/jamaah/:id" element={<JamaahDetail />} />
                <Route path="/paket" element={<PaketList />} />
                <Route path="/admin" element={<AdminList />} />
                <Route path="/riwayat" element={<AuditLog />} />
                <Route path="/pengumuman" element={<PengumumanList />} />
                <Route path="/manasik" element={<Manasik />} />
                <Route path="/manasik/sesi/:id" element={<SesiManasik />} />
              </Routes>
            </Layout>
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}
