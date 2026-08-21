import { useEffect, useRef } from 'react';
import { network } from '../../wailsjs/go/models';
import AdapterModeBadge from './AdapterModeBadge';

interface AdaptersModalProps {
  adapters: network.Adapter[];
  isLoading: boolean;
  isRefreshing: boolean;
  onToggleAdapter: (adapter: network.Adapter) => void;
  onRefresh: () => void;
  onClose: () => void;
}

const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export default function AdaptersModal({
  adapters,
  isLoading,
  isRefreshing,
  onToggleAdapter,
  onRefresh,
  onClose,
}: AdaptersModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);

  // Move focus into the dialog, keep Tab inside it, close on Escape and
  // restore focus to the trigger on unmount.
  useEffect(() => {
    const previouslyFocused = document.activeElement as HTMLElement | null;
    const node = dialogRef.current;
    node?.focus();

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onClose();
        return;
      }
      if (event.key !== 'Tab' || !node) return;

      const focusables = Array.from(node.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR));
      if (focusables.length === 0) {
        event.preventDefault();
        return;
      }
      const first = focusables[0];
      const last = focusables[focusables.length - 1];
      const active = document.activeElement as HTMLElement | null;

      if (event.shiftKey) {
        if (!active || active === first || active === node || !node.contains(active)) {
          event.preventDefault();
          last.focus();
        }
      } else if (active === last || !active || !node.contains(active)) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('keydown', onKeyDown);
      previouslyFocused?.focus?.();
    };
  }, [onClose]);

  return (
    <div className="absolute inset-0 z-50 flex items-center justify-center p-4 bg-black/50 dark:bg-black/70 backdrop-blur-sm">
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="adapters-modal-title"
        tabIndex={-1}
        className="w-full max-w-md bg-white dark:bg-gray-800 rounded-xl shadow-2xl p-5 flex flex-col max-h-[85vh] outline-none"
      >
        <div className="flex justify-between items-center mb-3">
          <h2 id="adapters-modal-title" className="text-base font-bold text-gray-800 dark:text-white">
            Tüm Ağ Adaptörleri
          </h2>
          <div className="flex items-center space-x-1">
            <button
              onClick={onRefresh}
              disabled={isRefreshing}
              className="text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 text-sm p-1 transition disabled:opacity-50"
              title="Adaptör Listesini Yenile"
              aria-label="Adaptör Listesini Yenile"
            >
              <span aria-hidden="true" className={`inline-block ${isRefreshing ? 'animate-spin' : ''}`}>
                🔄
              </span>
            </button>
            <button
              onClick={onClose}
              className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 text-sm p-1"
              title="Kapat"
              aria-label="Kapat"
            >
              <span aria-hidden="true">✕</span>
            </button>
          </div>
        </div>

        <div className="overflow-y-auto flex-1 pr-1 space-y-2.5">
          {adapters.map((a, idx) => (
            <div
              key={idx}
              className="p-3 border border-gray-200 dark:border-gray-700 rounded-lg flex flex-col space-y-1.5 bg-gray-50 dark:bg-gray-700/40"
            >
              <div className="flex justify-between items-center">
                <span className="font-semibold text-xs text-gray-800 dark:text-gray-200 select-text">
                  {a.name}
                </span>
                <div className="flex items-center gap-1.5">
                  <AdapterModeBadge mode={a.mode} />
                  <span
                    className={`text-[10px] px-2 py-0.5 rounded-full font-medium ${
                      a.enabled
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/60 dark:text-green-300'
                        : 'bg-red-100 text-red-700 dark:bg-red-900/60 dark:text-red-300'
                    }`}
                  >
                    {a.enabled ? 'Etkin' : 'Devre Dışı'}
                  </span>
                </div>
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-400 font-mono select-text">
                IP: {a.ipAddress || 'Yok'}
              </div>
              <div className="flex justify-end pt-1">
                <button
                  onClick={() => onToggleAdapter(a)}
                  disabled={isLoading}
                  className={`text-xs px-2.5 py-1 rounded font-medium transition ${
                    a.enabled
                      ? 'bg-red-100 text-red-600 hover:bg-red-200 dark:bg-red-900/40 dark:text-red-300 dark:hover:bg-red-800/60'
                      : 'bg-green-100 text-green-600 hover:bg-green-200 dark:bg-green-900/40 dark:text-green-300 dark:hover:bg-green-800/60'
                  }`}
                >
                  {a.enabled ? 'Devre Dışı Bırak' : 'Etkinleştir'}
                </button>
              </div>
            </div>
          ))}
          {adapters.length === 0 && (
            <div className="text-center text-gray-500 py-4 text-xs">Adaptör bulunamadı.</div>
          )}
        </div>
      </div>
    </div>
  );
}
