const BASE = '/api';

export async function fetchJamaah(paket = '') {
  const url = paket ? `${BASE}/jamaah?paket=${paket}` : `${BASE}/jamaah`;
  const res = await fetch(url);
  if (!res.ok) throw new Error('Gagal memuat data jamaah');
  return res.json();
}

export async function fetchJamaahById(id) {
  const res = await fetch(`${BASE}/jamaah/${id}`);
  if (!res.ok) throw new Error('Jamaah tidak ditemukan');
  return res.json();
}

export async function createJamaah(data) {
  const res = await fetch(`${BASE}/jamaah`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal membuat jamaah');
  }
  return res.json();
}

export async function updateJamaah(id, data) {
  const res = await fetch(`${BASE}/jamaah/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal memperbarui jamaah');
  }
  return res.json();
}

export async function deleteJamaah(id) {
  const res = await fetch(`${BASE}/jamaah/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Gagal menghapus jamaah');
  return res.json();
}

export async function uploadDocument(id, docType, file) {
  const formData = new FormData();
  formData.append('file', file);
  const res = await fetch(`${BASE}/jamaah/${id}/upload/${docType}`, {
    method: 'POST',
    body: formData,
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal upload dokumen');
  }
  return res.json();
}

export function getDocumentUrl(id, docType) {
  return `${BASE}/jamaah/${id}/dokumen/${docType}`;
}
