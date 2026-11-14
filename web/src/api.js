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

export async function fetchJamaah({ paket_id = '', search = '', status_pembayaran = '', kelengkapan = '', page = 1, limit = 10 } = {}) {
  const params = new URLSearchParams();
  if (paket_id) params.set('paket_id', paket_id);
  if (search) params.set('search', search);
  if (status_pembayaran) params.set('status_pembayaran', status_pembayaran);
  if (kelengkapan) params.set('kelengkapan', kelengkapan);
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

// Public: jamaah self-service manasik room lookup (no auth)
export async function cekKamarManasik(nik, tanggal_lahir) {
  const res = await fetch(`${BASE}/portal/cek-kamar`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nik, tanggal_lahir }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    if (res.status === 429) throw new Error('Terlalu banyak percobaan, silakan coba lagi beberapa menit lagi');
    throw new Error(data.error || 'Gagal memuat data');
  }
  return data;
}

// Bulk set manasik placement. Omit a field (undefined) to leave it unchanged; '' clears it.
export async function bulkSetManasik(ids, { gedung_manasik, kamar_manasik, rombongan, regu }) {
  const res = await authFetch(`${BASE}/jamaah/bulk-manasik`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ids, gedung_manasik, kamar_manasik, rombongan, regu }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Gagal mengatur kamar manasik');
  return data;
}

// Pengumuman (announcements)
export async function fetchPublicPengumuman() {
  const res = await fetch(`${BASE}/portal/pengumuman`);
  if (!res.ok) throw new Error('Gagal memuat pengumuman');
  return res.json();
}

export async function fetchPengumumanList() {
  const res = await authFetch(`${BASE}/pengumuman`);
  if (!res.ok) throw new Error('Gagal memuat pengumuman');
  return res.json();
}

async function savePengumuman(url, method, data) {
  const res = await authFetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      judul: data.judul,
      isi: data.isi,
      dipublikasikan: data.dipublikasikan,
      disematkan: data.disematkan,
    }),
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || 'Gagal menyimpan pengumuman');
  return body;
}

export function createPengumuman(data) {
  return savePengumuman(`${BASE}/pengumuman`, 'POST', data);
}

export function updatePengumuman(id, data) {
  return savePengumuman(`${BASE}/pengumuman/${id}`, 'PUT', data);
}

export async function deletePengumuman(id) {
  const res = await authFetch(`${BASE}/pengumuman/${id}`, { method: 'DELETE' });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || 'Gagal menghapus pengumuman');
  return body;
}

// Manasik (rooming list, import/export, attendance)
async function jsonOrThrow(res, fallback) {
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || fallback);
  return body;
}

export async function fetchManasikJamaah(paketId) {
  const res = await authFetch(`${BASE}/manasik/jamaah?paket_id=${paketId}`);
  return jsonOrThrow(res, 'Gagal memuat daftar kamar');
}

export async function downloadManasikExcel(paketId) {
  const res = await authFetch(`${BASE}/manasik/export?paket_id=${paketId}`);
  if (!res.ok) await jsonOrThrow(res, 'Gagal mengunduh excel');
  const disposition = res.headers.get('Content-Disposition') || '';
  const match = disposition.match(/filename="([^"]+)"/);
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = match ? match[1] : 'daftar-kamar-manasik.xlsx';
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export async function importManasik(paketId, file, dryRun) {
  const formData = new FormData();
  formData.append('file', file);
  const res = await authFetch(`${BASE}/manasik/import?paket_id=${paketId}&dry_run=${dryRun}`, {
    method: 'POST',
    body: formData,
  });
  return jsonOrThrow(res, 'Gagal import file');
}

export async function fetchSesiManasik(paketId) {
  const res = await authFetch(`${BASE}/manasik/sesi?paket_id=${paketId}`);
  return jsonOrThrow(res, 'Gagal memuat sesi manasik');
}

export async function createSesiManasik(data) {
  const res = await authFetch(`${BASE}/manasik/sesi`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return jsonOrThrow(res, 'Gagal membuat sesi');
}

export async function deleteSesiManasik(id) {
  const res = await authFetch(`${BASE}/manasik/sesi/${id}`, { method: 'DELETE' });
  return jsonOrThrow(res, 'Gagal menghapus sesi');
}

export async function fetchSesiDetail(id) {
  const res = await authFetch(`${BASE}/manasik/sesi/${id}`);
  return jsonOrThrow(res, 'Gagal memuat sesi');
}

export async function markHadir(sesiId, payload) {
  const res = await authFetch(`${BASE}/manasik/sesi/${sesiId}/hadir`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return jsonOrThrow(res, 'Gagal mencatat kehadiran');
}

export async function unmarkHadir(sesiId, jamaahId) {
  const res = await authFetch(`${BASE}/manasik/sesi/${sesiId}/hadir/${jamaahId}`, { method: 'DELETE' });
  return jsonOrThrow(res, 'Gagal membatalkan kehadiran');
}

// Tagihan (outstanding balance report)
function tagihanParams({ paket_id = '', belum_lunas = false } = {}) {
  const params = new URLSearchParams();
  if (paket_id) params.set('paket_id', paket_id);
  if (belum_lunas) params.set('belum_lunas', 'true');
  return params;
}

export async function fetchTagihan(filters) {
  const res = await authFetch(`${BASE}/tagihan?${tagihanParams(filters)}`);
  return jsonOrThrow(res, 'Gagal memuat laporan tagihan');
}

export async function downloadTagihanExcel(filters) {
  const res = await authFetch(`${BASE}/tagihan/export?${tagihanParams(filters)}`);
  if (!res.ok) await jsonOrThrow(res, 'Gagal mengunduh excel');
  const match = (res.headers.get('Content-Disposition') || '').match(/filename="([^"]+)"/);
  const url = URL.createObjectURL(await res.blob());
  const a = document.createElement('a');
  a.href = url;
  a.download = match ? match[1] : 'laporan-tagihan.xlsx';
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}
