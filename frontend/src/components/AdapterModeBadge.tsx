interface AdapterModeBadgeProps {
  mode: string;
  className?: string;
}

const LABELS: Record<string, string> = {
  dhcp: 'DHCP',
  manual: 'Statik',
};

const STYLES: Record<string, string> = {
  dhcp: 'bg-blue-100 text-blue-700 dark:bg-blue-900/60 dark:text-blue-300',
  manual: 'bg-amber-100 text-amber-700 dark:bg-amber-900/60 dark:text-amber-300',
};

const TITLES: Record<string, string> = {
  dhcp: 'Adres otomatik olarak alınıyor (DHCP)',
  manual: 'Adres elle ayarlanmış (statik)',
};

/**
 * Shows how the adapter currently gets its address. The backend leaves the mode
 * empty when it could not determine it, in which case nothing is rendered
 * rather than a misleading guess.
 */
export default function AdapterModeBadge({ mode, className = '' }: AdapterModeBadgeProps) {
  const label = LABELS[mode];
  if (!label) return null;

  return (
    <span
      className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${STYLES[mode]} ${className}`}
      title={TITLES[mode]}
    >
      {label}
    </span>
  );
}
