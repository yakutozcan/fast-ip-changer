import { diagnostics } from '../../wailsjs/go/models';

interface ConnectionStatusBarProps {
  quickStatus: diagnostics.QuickCheckResult | null;
  isCheckingStatus: boolean;
  onRefresh: () => void;
  checkPublicIP: boolean;
  onToggleCheckPublicIP: (next: boolean) => void;
}

export default function ConnectionStatusBar({
  quickStatus,
  isCheckingStatus,
  onRefresh,
  checkPublicIP,
  onToggleCheckPublicIP,
}: ConnectionStatusBarProps) {
  return (
    <div className="mb-4 bg-gray-50 dark:bg-gray-700/50 border border-gray-200 dark:border-gray-700 rounded-lg p-2.5 flex flex-wrap items-center justify-between gap-y-1.5 text-xs">
      <div className="flex items-center space-x-3">
        <div className="flex items-center space-x-1.5" title="Ağ Geçidi (Gateway)">
          <span
            className={`w-2 h-2 rounded-full ${quickStatus?.gatewayOk ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}
          ></span>
          <span className="text-gray-500 dark:text-gray-400">GW:</span>
          <span className="font-semibold text-gray-700 dark:text-gray-200 select-text">
            {quickStatus?.gatewayOk ? quickStatus.gatewayLatency : 'Yok'}
          </span>
        </div>
        <div className="flex items-center space-x-1.5" title="İnternet (1.1.1.1)">
          <span
            className={`w-2 h-2 rounded-full ${quickStatus?.internetOk ? 'bg-green-500' : 'bg-red-500'}`}
          ></span>
          <span className="text-gray-500 dark:text-gray-400">Net:</span>
          <span className="font-semibold text-gray-700 dark:text-gray-200 select-text">
            {quickStatus?.internetOk ? quickStatus.internetLatency : 'Yok'}
          </span>
        </div>
      </div>

      <div className="flex items-center space-x-2">
        {quickStatus?.publicIp && (
          <span
            className="text-gray-500 dark:text-gray-400 font-mono select-text"
            title="Dış IP (Public IP)"
          >
            {quickStatus.publicIp}
          </span>
        )}

        <span className="flex items-center space-x-1">
          <input
            id="check-public-ip"
            type="checkbox"
            className="w-3 h-3 accent-blue-600"
            checked={checkPublicIP}
            onChange={(e) => onToggleCheckPublicIP(e.target.checked)}
          />
          <label
            htmlFor="check-public-ip"
            className="text-[10px] text-gray-500 dark:text-gray-400 cursor-pointer"
            title="Açıkken dış IP adresiniz harici bir servise (api.ipify.org) sorulur. Kapalıyken hiçbir dış istek yapılmaz."
          >
            Dış IP sorgusu
          </label>
        </span>

        <button
          onClick={onRefresh}
          disabled={isCheckingStatus}
          className="text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition"
          title="Bağlantıyı Yeniden Test Et"
          aria-label="Bağlantıyı Yeniden Test Et"
        >
          <span aria-hidden="true" className={`inline-block ${isCheckingStatus ? 'animate-spin' : ''}`}>
            🔄
          </span>
        </button>
      </div>
    </div>
  );
}
