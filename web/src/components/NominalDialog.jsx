import { useEffect, useMemo, useRef, useState } from 'react';
import RupiahInput from './RupiahInput';
import { formatRupiah } from '../utils/format';

// Asks for the payment amount when uploading a DP / pelunasan receipt.
// Shows the amount formatted (3.000.000) and in words while typing.
export default function NominalDialog({ title, file, sisa, onCancel, onSubmit }) {
  const [nominal, setNominal] = useState(0);
  const [busy, setBusy] = useState(false);
  const inputRef = useRef(null);

  const previewUrl = useMemo(
    () => (file && file.type.startsWith('image/') ? URL.createObjectURL(file) : ''),
    [file],
  );
  useEffect(() => () => previewUrl && URL.revokeObjectURL(previewUrl), [previewUrl]);
  useEffect(() => {
    inputRef.current?.focus();
    const onKey = (e) => e.key === 'Escape' && !busy && onCancel();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [busy, onCancel]);

  const submit = async (e) => {
    e.preventDefault();
    if (nominal <= 0 || busy) return;
    setBusy(true);
    try {
      await onSubmit(nominal);
    } finally {
      setBusy(false);
    }
  };

  const lebih = sisa > 0 && nominal > sisa;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => !busy && onCancel()}>
      <form
        role="dialog"
        aria-modal="true"
        aria-labelledby="nominal-title"
        onSubmit={submit}
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-md bg-white rounded-lg shadow-xl p-6"
      >
        <h2 id="nominal-title" className="text-lg font-semibold text-gray-900">
          {title}
        </h2>

        <div className="mt-3 flex items-center gap-3 rounded-md bg-gray-50 border border-gray-200 p-2">
          {previewUrl ? (
            <img src={previewUrl} alt="" className="h-14 w-14 object-cover rounded" />
          ) : (
            <div className="h-14 w-14 rounded bg-gray-200 flex items-center justify-center text-xs font-medium text-gray-600">
              PDF
            </div>
          )}
          <p className="text-sm text-gray-700 break-all">{file?.name}</p>
        </div>

        <label htmlFor="nominal-input" className="mt-4 block text-sm font-medium text-gray-700">
          Nominal pembayaran
        </label>
        <div className="mt-1">
          <RupiahInput
            id="nominal-input"
            ref={inputRef}
            value={nominal}
            onChange={setNominal}
            placeholder="0"
            className="block w-full rounded-md border border-gray-300 px-3 py-2.5 text-lg font-semibold tabular-nums shadow-sm focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          />
        </div>

        {sisa > 0 && (
          <div className="mt-2 flex items-center justify-between gap-2 text-sm">
            <span className="text-gray-500">Sisa tagihan: {formatRupiah(sisa)}</span>
            <button
              type="button"
              onClick={() => setNominal(sisa)}
              className="text-emerald-700 hover:text-emerald-900 font-medium"
            >
              Isi sisa tagihan
            </button>
          </div>
        )}
        {lebih && (
          <p className="mt-2 text-sm text-amber-700">Nominal melebihi sisa tagihan ({formatRupiah(sisa)}).</p>
        )}

        <div className="mt-6 flex justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            disabled={busy}
            className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-200 disabled:opacity-50"
          >
            Batal
          </button>
          <button
            type="submit"
            disabled={nominal <= 0 || busy}
            className="px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-md hover:bg-emerald-700 disabled:opacity-50"
          >
            {busy ? 'Mengupload...' : 'Upload'}
          </button>
        </div>
      </form>
    </div>
  );
}
