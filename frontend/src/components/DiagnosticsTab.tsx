interface DiagnosticsTabProps {
  targetHost: string;
  onChangeTargetHost: (value: string) => void;
  pingCount: number;
  onChangePingCount: (value: number) => void;
  gateway: string;
  consoleOutput: string;
  isExecutingTool: boolean;
  isCancelling: boolean;
  onPing: () => void;
  onTraceRoute: () => void;
  onCancel: () => void;
  onCopyOutput: () => void;
  onClearOutput: () => void;
}

const PILL_CLASSES =
  'text-[11px] bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 px-2 py-0.5 rounded text-gray-700 dark:text-gray-300';

export default function DiagnosticsTab({
  targetHost,
  onChangeTargetHost,
  pingCount,
  onChangePingCount,
  gateway,
  consoleOutput,
  isExecutingTool,
  isCancelling,
  onPing,
  onTraceRoute,
  onCancel,
  onCopyOutput,
  onClearOutput,
}: DiagnosticsTabProps) {
  return (
    <div className="flex flex-col space-y-3">
      <div>
        <label
          htmlFor="target-host"
          className="block text-xs font-semibold text-gray-600 dark:text-gray-300 mb-1"
        >
          Hedef IP veya Alan Adı
        </label>
        <div className="flex space-x-2">
          <input
            id="target-host"
            type="text"
            placeholder="Örn: 192.168.1.1 veya 8.8.8.8"
            autoComplete="off"
            spellCheck={false}
            className="flex-1 border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm p-2 border text-xs font-mono select-text"
            value={targetHost}
            onChange={(e) => onChangeTargetHost(e.target.value)}
          />
          <select
            id="ping-count"
            value={pingCount}
            onChange={(e) => onChangePingCount(Number(e.target.value))}
            className="border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md shadow-sm p-2 border text-xs"
            title="Ping Adedi"
            aria-label="Ping Adedi"
          >
            <option value={1}>1 Paket</option>
            <option value={4}>4 Paket</option>
            <option value={10}>10 Paket</option>
          </select>
        </div>

        {/* Quick Fill Pills */}
        <div className="flex flex-wrap gap-1.5 mt-2">
          {gateway && (
            <button onClick={() => onChangeTargetHost(gateway)} className={`${PILL_CLASSES} font-mono`}>
              GW ({gateway})
            </button>
          )}
          <button onClick={() => onChangeTargetHost('1.1.1.1')} className={`${PILL_CLASSES} font-mono`}>
            1.1.1.1 (Cloudflare)
          </button>
          <button onClick={() => onChangeTargetHost('8.8.8.8')} className={`${PILL_CLASSES} font-mono`}>
            8.8.8.8 (Google)
          </button>
          <button onClick={() => onChangeTargetHost('google.com')} className={PILL_CLASSES}>
            google.com
          </button>
        </div>
      </div>

      {/* Diagnostic Action Buttons */}
      <div className="grid grid-cols-2 gap-2 pt-1">
        <button
          onClick={onPing}
          disabled={isExecutingTool}
          className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-3 rounded-md shadow transition text-xs flex items-center justify-center space-x-1.5 disabled:opacity-50"
        >
          <span aria-hidden="true">⚡</span>
          <span>{isExecutingTool ? 'Ping Atılıyor...' : 'Ping Gönder'}</span>
        </button>
        <button
          onClick={onTraceRoute}
          disabled={isExecutingTool}
          className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-3 rounded-md shadow transition text-xs flex items-center justify-center space-x-1.5 disabled:opacity-50"
        >
          <span aria-hidden="true">🧭</span>
          <span>{isExecutingTool ? 'İz Sürülüyor...' : 'Traceroute Başlat'}</span>
        </button>
        {isExecutingTool && (
          <button
            onClick={onCancel}
            disabled={isCancelling}
            className="col-span-2 bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/40 dark:text-red-300 dark:hover:bg-red-800/60 font-semibold py-2 px-3 rounded-md transition text-xs flex items-center justify-center space-x-1.5 disabled:opacity-50"
          >
            <span aria-hidden="true">⛔</span>
            <span>{isCancelling ? 'İptal Ediliyor...' : 'İptal'}</span>
          </button>
        )}
      </div>

      {/* Terminal Output Console */}
      <div className="relative mt-2">
        <div className="flex justify-between items-center bg-gray-900 text-gray-400 px-3 py-1.5 rounded-t-md text-[11px] font-mono border-b border-gray-800">
          <span>TERMINAL ÇIKTISI</span>
          <div className="flex space-x-2">
            <button
              onClick={onCopyOutput}
              className="hover:text-white transition"
              title="Panoya Kopyala"
              aria-label="Terminal çıktısını panoya kopyala"
            >
              <span aria-hidden="true">📋</span> Kopyala
            </button>
            <button
              onClick={onClearOutput}
              className="hover:text-white transition"
              title="Temizle"
              aria-label="Terminal çıktısını temizle"
            >
              <span aria-hidden="true">🗑️</span> Temizle
            </button>
          </div>
        </div>
        <pre
          data-native-menu="true"
          tabIndex={0}
          aria-label="Terminal çıktısı"
          className="bg-gray-950 text-green-400 p-3 rounded-b-md text-xs font-mono h-48 overflow-y-auto whitespace-pre-wrap select-text cursor-text selection:bg-green-900 selection:text-white border border-gray-900"
        >
          {consoleOutput || '// Ping veya Traceroute sonuçları burada listelenecektir...'}
        </pre>
      </div>
    </div>
  );
}
