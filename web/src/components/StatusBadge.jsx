const statusConfig = {
  lunas: { label: 'Lunas', className: 'bg-green-100 text-green-800' },
  dp: { label: 'DP', className: 'bg-yellow-100 text-yellow-800' },
  belum_bayar: { label: 'Belum Bayar', className: 'bg-red-100 text-red-800' },
};

export default function StatusBadge({ status }) {
  const config = statusConfig[status] || {
    label: status,
    className: 'bg-gray-100 text-gray-800',
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.className}`}
    >
      {config.label}
    </span>
  );
}
