import { forwardRef } from 'react';
import { terbilang as toWords } from '../utils/format';

// Money input: shows thousands separators while typing (3000000 -> 3.000.000),
// reports a plain number, and by default spells the amount out below
// ("tiga juta rupiah") so a missing/extra zero is easy to spot.
const RupiahInput = forwardRef(function RupiahInput(
  { value, onChange, className = '', terbilang = true, id, ...props },
  ref,
) {
  const display = value ? Number(value).toLocaleString('id-ID') : '';
  const words = terbilang ? toWords(value) : '';
  const wordsId = id ? `${id}-terbilang` : undefined;
  return (
    <div>
      <div className="relative">
        <span className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-sm text-gray-500">
          Rp
        </span>
        <input
          {...props}
          id={id}
          ref={ref}
          inputMode="numeric"
          autoComplete="off"
          value={display}
          aria-describedby={words ? wordsId : undefined}
          onChange={(e) => onChange(Number(e.target.value.replace(/\D/g, '')) || 0)}
          className={`${className} pl-9`}
        />
      </div>
      {terbilang && (
        <p id={wordsId} className="mt-1 min-h-4 text-xs text-gray-500 first-letter:uppercase" aria-live="polite">
          {words}
        </p>
      )}
    </div>
  );
});

export default RupiahInput;
