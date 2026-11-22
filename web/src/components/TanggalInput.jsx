import { useState } from 'react';

const BULAN = [
  'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
  'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
];

const selectClass =
  'w-full px-2 py-3 border border-gray-300 rounded-md text-base bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500';

function parse(value) {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(value || '');
  return m ? { y: m[1], m: String(Number(m[2])), d: String(Number(m[3])) } : { y: '', m: '', d: '' };
}

// Day/month/year dropdowns; easier than the native date picker on Android
// for dates far in the past. value/onChange use yyyy-mm-dd ('' when incomplete).
export default function TanggalInput({ value, onChange, required = false, minYear = 1900, maxYear = new Date().getFullYear() }) {
  const [parts, setParts] = useState(() => parse(value));
  const [lastValue, setLastValue] = useState(value);

  // Resync when the parent loads or resets the value (e.g. editing an existing jamaah)
  if (value !== lastValue) {
    setLastValue(value);
    setParts(parse(value));
  }

  function emit(p) {
    if (!p.y || !p.m || !p.d) return '';
    return `${p.y}-${p.m.padStart(2, '0')}-${p.d.padStart(2, '0')}`;
  }

  function update(key, v) {
    const next = { ...parts, [key]: v };
    if (next.y && next.m && next.d) {
      const maxDay = new Date(Number(next.y), Number(next.m), 0).getDate();
      if (Number(next.d) > maxDay) next.d = String(maxDay);
    }
    setParts(next);
    const out = emit(next);
    setLastValue(out);
    onChange(out);
  }

  const daysInMonth = parts.y && parts.m ? new Date(Number(parts.y), Number(parts.m), 0).getDate() : 31;
  const years = [];
  for (let y = maxYear; y >= minYear; y--) years.push(y);

  return (
    <div className="grid grid-cols-[1fr_2fr_1.3fr] gap-2">
      <select value={parts.d} onChange={(e) => update('d', e.target.value)} required={required} className={selectClass} aria-label="Tanggal">
        <option value="">Tgl</option>
        {Array.from({ length: daysInMonth }, (_, i) => i + 1).map((d) => (
          <option key={d} value={d}>{d}</option>
        ))}
      </select>
      <select value={parts.m} onChange={(e) => update('m', e.target.value)} required={required} className={selectClass} aria-label="Bulan">
        <option value="">Bulan</option>
        {BULAN.map((b, i) => (
          <option key={b} value={i + 1}>{b}</option>
        ))}
      </select>
      <select value={parts.y} onChange={(e) => update('y', e.target.value)} required={required} className={selectClass} aria-label="Tahun">
        <option value="">Tahun</option>
        {years.map((y) => (
          <option key={y} value={y}>{y}</option>
        ))}
      </select>
    </div>
  );
}
