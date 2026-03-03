import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import JamaahList from './pages/JamaahList';
import JamaahForm from './pages/JamaahForm';
import JamaahDetail from './pages/JamaahDetail';

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<JamaahList />} />
        <Route path="/jamaah/new" element={<JamaahForm />} />
        <Route path="/jamaah/:id/edit" element={<JamaahForm />} />
        <Route path="/jamaah/:id" element={<JamaahDetail />} />
      </Routes>
    </Layout>
  );
}
