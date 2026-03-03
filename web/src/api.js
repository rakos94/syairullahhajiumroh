const BASE = '/api';

export async function fetchJamaah({ paket_id = '', page = 1, limit = 10 } = {}) {
  const params = new URLSearchParams();
  if (paket_id) params.set('paket_id', paket_id);
  params.set('page', page);
  params.set('limit', limit);
  const res = await fetch(`${BASE}/jamaah?${params}`);
  if (!res.ok) throw new Error('Gagal memuat data jamaah');
  return res.json();
}

export async function fetchPaketList(tipe = '') {
  const params = new URLSearchParams();
  if (tipe) params.set('tipe', tipe);
  const res = await fetch(`${BASE}/paket?${params}`);
  if (!res.ok) throw new Error('Gagal memuat data paket');
  return res.json();
}

export async function createPaket(data) {
  const res = await fetch(`${BASE}/paket`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal membuat paket');
  }
  return res.json();
}

export async function updatePaket(id, data) {
  const res = await fetch(`${BASE}/paket/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal memperbarui paket');
  }
  return res.json();
}

export async function deletePaket(id) {
  const res = await fetch(`${BASE}/paket/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Gagal menghapus paket');
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

export async function uploadDocument(id, docType, file, nominal) {
  const formData = new FormData();
  formData.append('file', file);
  if (nominal !== undefined) formData.append('nominal', nominal);
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

export function getDocumentDownloadUrl(id, docType) {
  return `${BASE}/jamaah/${id}/dokumen/${docType}?download=true`;
}

export function getMultiDocumentUrl(id, docType, index) {
  return `${BASE}/jamaah/${id}/dokumen/${docType}?index=${index}`;
}

export function getMultiDocumentDownloadUrl(id, docType, index) {
  return `${BASE}/jamaah/${id}/dokumen/${docType}?index=${index}&download=true`;
}

export async function deleteDocument(id, docType, index) {
  const res = await fetch(`${BASE}/jamaah/${id}/dokumen/${docType}?index=${index}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal menghapus dokumen');
  }
  return res.json();
}
