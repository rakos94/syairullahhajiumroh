import { useEffect, useState } from 'react';
import QRCode from 'qrcode';

// Personal attendance QR shown to the jamaah on the Cek Kamar result.
export default function ManasikQr({ token, nama }) {
  const [dataUrl, setDataUrl] = useState('');

  useEffect(() => {
    let cancelled = false;
    QRCode.toDataURL(token, { width: 480, margin: 2, errorCorrectionLevel: 'M' })
      .then((url) => !cancelled && setDataUrl(url))
      .catch(() => !cancelled && setDataUrl(''));
    return () => {
      cancelled = true;
    };
  }, [token]);

  const filename = `qr-manasik-${(nama || 'jamaah').toLowerCase().replace(/[^a-z0-9]+/g, '-')}.png`;

  return (
    <div className="mt-6 rounded-lg border border-gray-200 p-4 text-center">
      <p className="text-sm font-semibold text-gray-900">QR Absensi Manasik</p>
      <p className="text-xs text-gray-500 mt-1">
        Tunjukkan QR ini kepada petugas saat registrasi setiap sesi manasik.
      </p>
      {dataUrl ? (
        <>
          <img
            src={dataUrl}
            alt={`QR absensi manasik ${nama}`}
            className="mx-auto mt-3 w-56 h-56 [image-rendering:pixelated]"
          />
          <a
            href={dataUrl}
            download={filename}
            className="mt-2 inline-block text-sm font-medium text-emerald-700 hover:underline"
          >
            Simpan gambar QR
          </a>
        </>
      ) : (
        <div className="mx-auto mt-3 w-56 h-56 bg-gray-100 rounded animate-pulse" aria-hidden="true" />
      )}
    </div>
  );
}
