import { isWindows } from '../lib/platform';

/**
 * Persistent warning shown when the app is not running with administrator
 * rights. Never blocks the UI — diagnostics work unprivileged.
 */
export default function ElevationBanner() {
  const windowsHost = isWindows();

  return (
    <div
      role="alert"
      className="mb-4 flex items-start space-x-2 rounded-lg border border-amber-300 bg-amber-50 p-2.5 text-[11px] leading-snug text-amber-800 dark:border-amber-600/60 dark:bg-amber-900/40 dark:text-amber-200"
    >
      <span aria-hidden="true" className="text-sm leading-none">
        🔒
      </span>
      <div className="space-y-0.5 select-text">
        <p className="font-semibold">Yönetici yetkisi yok</p>
        <p>
          {windowsHost
            ? 'Ağ ayarlarını değiştirmek için uygulamayı yönetici olarak yeniden başlatın (sağ tık → Yönetici olarak çalıştır).'
            : 'Ağ ayarlarını değiştirmek yönetici yetkisi gerektirir. Ayarları uygularken sistem şifreniz istenecektir.'}
        </p>
        <p className="opacity-80">Ping ve Traceroute araçları yetki gerektirmez.</p>
      </div>
    </div>
  );
}
