// Text input that shows thousands separators (25.000.000) but reports a number.
export default function RupiahInput({ value, onChange, className = '', ...props }) {
  const display = value ? Number(value).toLocaleString('id-ID') : '';
  return (
    <div className="relative">
      <span className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-sm text-gray-500">
        Rp
      </span>
      <input
        {...props}
        inputMode="numeric"
        value={display}
        onChange={(e) => onChange(Number(e.target.value.replace(/\D/g, '')) || 0)}
        className={`${className} pl-9`}
      />
    </div>
  );
}
