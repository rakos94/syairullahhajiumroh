import imageCompression from 'browser-image-compression';

const BASE = '/api';

async function authFetch(url, options = {}) {
  const token = localStorage.getItem('token');
  const headers = { ...options.headers };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(url, { ...options, headers });
  if (res.status === 401) {
    localStorage.removeItem('token');
    window.location.href = '/login';
    throw new Error('Sesi berakhir, silakan login kembali');
  }
  return res;
}

export async function fetchJamaah({ paket_id = '', search = '', status_pembayaran = '', page = 1, limit = 10 } = {}) {
  const params = new URLSearchParams();
  if (paket_id) params.set('paket_id', paket_id);
  if (search) params.set('search', search);
  if (status_pembayaran) params.set('status_pembayaran', status_pembayaran);
  params.set('page', page);
  params.set('limit', limit);
  const res = await authFetch(`${BASE}/jamaah?${params}`);
  if (!res.ok) throw new Error('Gagal memuat data jamaah');
  return res.json();
}

export async function fetchPaketList(tipe = '') {
  const params = new URLSearchParams();
  if (tipe) params.set('tipe', tipe);
  const res = await authFetch(`${BASE}/paket?${params}`);
  if (!res.ok) throw new Error('Gagal memuat data paket');
  return res.json();
}

export async function createPaket(data) {
  const res = await authFetch(`${BASE}/paket`, {
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
  const res = await authFetch(`${BASE}/paket/${id}`, {
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
  const res = await authFetch(`${BASE}/paket/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Gagal menghapus paket');
  return res.json();
}

export async function fetchJamaahById(id) {
  const res = await authFetch(`${BASE}/jamaah/${id}`);
  if (!res.ok) throw new Error('Jamaah tidak ditemukan');
  return res.json();
}

export async function createJamaah(data) {
  const res = await authFetch(`${BASE}/jamaah`, {
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
  const res = await authFetch(`${BASE}/jamaah/${id}`, {
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
  const res = await authFetch(`${BASE}/jamaah/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Gagal menghapus jamaah');
  return res.json();
}

export async function uploadDocument(id, docType, file, nominal) {
  if (file.type.startsWith('image/')) {
    file = await imageCompression(file, {
      maxSizeMB: 1,
      maxWidthOrHeight: 1920,
      useWebWorker: true,
    });
  }
  const formData = new FormData();
  formData.append('file', file);
  if (nominal !== undefined) formData.append('nominal', nominal);
  const res = await authFetch(`${BASE}/jamaah/${id}/upload/${docType}`, {
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
  const res = await authFetch(`${BASE}/jamaah/${id}/dokumen/${docType}?index=${index}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal menghapus dokumen');
  }
  return res.json();
}

// Admin management
export async function fetchAdminList() {
  const res = await authFetch(`${BASE}/admin`);
  if (!res.ok) throw new Error('Gagal memuat data admin');
  return res.json();
}

export async function createAdmin(data) {
  const res = await authFetch(`${BASE}/admin`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal membuat admin');
  }
  return res.json();
}

export async function updateAdmin(id, data) {
  const res = await authFetch(`${BASE}/admin/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal memperbarui admin');
  }
  return res.json();
}

export async function deleteAdmin(id) {
  const res = await authFetch(`${BASE}/admin/${id}`, { method: 'DELETE' });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error || 'Gagal menghapus admin');
  }
  return res.json();
}

// Statistics
export async function fetchStatistics() {
  const res = await authFetch(`${BASE}/statistics`);
  if (!res.ok) throw new Error('Gagal memuat statistik');
  return res.json();
}

// Audit logs
export async function fetchAuditLogs({ entity_type = '', page = 1, limit = 20 } = {}) {
  const params = new URLSearchParams();
  if (entity_type) params.set('entity_type', entity_type);
  params.set('page', page);
  params.set('limit', limit);
  const res = await authFetch(`${BASE}/audit-logs?${params}`);
  if (!res.ok) throw new Error('Gagal memuat audit log');
  return res.json();
}

export async function fetchAuditLogsByEntity(entityType, entityId, { page = 1, limit = 20 } = {}) {
  const params = new URLSearchParams();
  params.set('page', page);
  params.set('limit', limit);
  const res = await authFetch(`${BASE}/audit-logs/${entityType}/${entityId}?${params}`);
  if (!res.ok) throw new Error('Gagal memuat audit log');
  return res.json();
}
