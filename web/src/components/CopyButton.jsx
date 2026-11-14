import { useEffect, useRef, useState } from 'react';

async function copyText(text) {
  // navigator.clipboard only exists on HTTPS/localhost; fall back for plain
  // http:// (e.g. opening the app via the office LAN IP).
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.setAttribute('readonly', '');
  ta.style.position = 'fixed';
  ta.style.opacity = '0';
  document.body.appendChild(ta);
  ta.select();
  const ok = document.execCommand('copy');
  ta.remove();
  if (!ok) throw new Error('copy failed');
}

// Small icon button that copies `text` and briefly shows a check mark.
export default function CopyButton({ text, label = 'Salin' }) {
  const [state, setState] = useState('idle'); // idle | copied | error
  const timer = useRef(null);

  useEffect(() => () => clearTimeout(timer.current), []);

  const onClick = async (e) => {
    e.stopPropagation();
    try {
      await copyText(text);
      setState('copied');
    } catch {
      setState('error');
    }
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setState('idle'), 1500);
  };

  const title = state === 'copied' ? 'Tersalin' : state === 'error' ? 'Gagal menyalin' : label;

  return (
    <button
      type="button"
      onClick={onClick}
      title={title}
      aria-label={`${label} ${text}`}
      className={`inline-flex items-center gap-1 rounded p-1 align-middle transition-colors ${
        state === 'copied' ? 'text-emerald-600' : state === 'error' ? 'text-red-600' : 'text-gray-400 hover:text-gray-700 hover:bg-gray-100'
      }`}
    >
      {state === 'copied' ? (
        <svg aria-hidden="true" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2" className="h-4 w-4">
          <path strokeLinecap="round" strokeLinejoin="round" d="M4 10.5l4 4 8-9" />
        </svg>
      ) : (
        <svg aria-hidden="true" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.6" className="h-4 w-4">
          <rect x="7" y="7" width="10" height="10" rx="2" />
          <path d="M13 7V5a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2" />
        </svg>
      )}
      <span className="text-xs" aria-live="polite">
        {state === 'copied' ? 'Tersalin' : state === 'error' ? 'Gagal' : ''}
      </span>
    </button>
  );
}
