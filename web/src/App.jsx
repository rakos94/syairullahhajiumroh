import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import JamaahList from './pages/JamaahList';
import JamaahForm from './pages/JamaahForm';
import JamaahDetail from './pages/JamaahDetail';
import PaketList from './pages/PaketList';
import AdminList from './pages/AdminList';
import AuditLog from './pages/AuditLog';
import Login from './pages/Login';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/*"
        element={
          <ProtectedRoute>
            <Layout>
              <Routes>
                <Route path="/" element={<JamaahList />} />
                <Route path="/jamaah/new" element={<JamaahForm />} />
                <Route path="/jamaah/:id/edit" element={<JamaahForm />} />
                <Route path="/jamaah/:id" element={<JamaahDetail />} />
                <Route path="/paket" element={<PaketList />} />
                <Route path="/admin" element={<AdminList />} />
                <Route path="/riwayat" element={<AuditLog />} />
              </Routes>
            </Layout>
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}
