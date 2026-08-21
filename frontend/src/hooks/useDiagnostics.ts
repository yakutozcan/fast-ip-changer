import { useCallback, useState } from 'react';
import {
  CancelDiagnostics,
  PingHost,
  TraceRouteHost,
} from '../../wailsjs/go/main/App';
import type { ShowMessage } from './useMessage';

/** Ping / traceroute state, including cancellation of an in-flight run. */
export function useDiagnostics(showMessage: ShowMessage) {
  const [targetHost, setTargetHost] = useState('');
  const [pingCount, setPingCount] = useState<number>(4);
  const [consoleOutput, setConsoleOutput] = useState<string>('');
  const [isExecutingTool, setIsExecutingTool] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);

  const requireHost = useCallback(() => {
    if (!targetHost.trim()) {
      showMessage('Lütfen geçerli bir hedef IP veya adres girin', 'error');
      return false;
    }
    return true;
  }, [showMessage, targetHost]);

  const runPing = useCallback(async () => {
    if (!requireHost()) return;
    setIsExecutingTool(true);
    setConsoleOutput(`> ping -c ${pingCount} ${targetHost} başlatılıyor...\n`);
    try {
      const res = await PingHost(targetHost.trim(), pingCount);
      setConsoleOutput((prev) => prev + res + '\n[Ping Tamamlandı]\n');
    } catch (err) {
      setConsoleOutput((prev) => prev + `Hata: ${String(err)}\n`);
    } finally {
      setIsExecutingTool(false);
      setIsCancelling(false);
    }
  }, [pingCount, requireHost, targetHost]);

  const runTraceRoute = useCallback(async () => {
    if (!requireHost()) return;
    setIsExecutingTool(true);
    setConsoleOutput(
      `> traceroute ${targetHost} başlatılıyor (bu işlem birkaç saniye sürebilir)...\n`,
    );
    try {
      const res = await TraceRouteHost(targetHost.trim());
      setConsoleOutput((prev) => prev + res + '\n[Traceroute Tamamlandı]\n');
    } catch (err) {
      setConsoleOutput((prev) => prev + `Hata: ${String(err)}\n`);
    } finally {
      setIsExecutingTool(false);
      setIsCancelling(false);
    }
  }, [requireHost, targetHost]);

  const cancel = useCallback(async () => {
    setIsCancelling(true);
    try {
      await CancelDiagnostics();
      setConsoleOutput((prev) => prev + '\n[İptal isteği gönderildi]\n');
    } catch (err) {
      setConsoleOutput((prev) => prev + `Hata: ${String(err)}\n`);
      setIsCancelling(false);
    }
  }, []);

  const clearOutput = useCallback(() => setConsoleOutput(''), []);

  const copyOutput = useCallback(async () => {
    if (!consoleOutput) return;
    try {
      await navigator.clipboard.writeText(consoleOutput);
      showMessage('Çıktı panoya kopyalandı', 'success');
    } catch (err) {
      console.error(err);
      showMessage('Çıktı kopyalanamadı', 'error');
    }
  }, [consoleOutput, showMessage]);

  return {
    targetHost,
    setTargetHost,
    pingCount,
    setPingCount,
    consoleOutput,
    isExecutingTool,
    isCancelling,
    runPing,
    runTraceRoute,
    cancel,
    clearOutput,
    copyOutput,
  };
}
