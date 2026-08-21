import { useCallback, useMemo, useState } from 'react';
import { GetAdapters } from '../../wailsjs/go/main/App';
import { network } from '../../wailsjs/go/models';
import type { ShowMessage } from './useMessage';

/** Network adapter list plus the currently selected adapter. */
export function useAdapters(showMessage: ShowMessage) {
  const [adapters, setAdapters] = useState<network.Adapter[]>([]);
  const [selectedAdapter, setSelectedAdapter] = useState<string>('');
  const [isRefreshing, setIsRefreshing] = useState(false);

  const reload = useCallback(async () => {
    setIsRefreshing(true);
    try {
      const data = await GetAdapters();
      const list = data || [];
      setAdapters(list);
      if (list.length > 0) {
        setSelectedAdapter((prev) => {
          if (prev && list.some((a) => a.name === prev)) return prev;
          const fallback = list.find((a) => a.enabled) || list[0];
          return fallback.name;
        });
      }
    } catch (err) {
      console.error(err);
      showMessage('Adaptörler yüklenemedi: ' + String(err), 'error');
    } finally {
      setIsRefreshing(false);
    }
  }, [showMessage]);

  const currentAdapter = useMemo(
    () => adapters.find((a) => a.name === selectedAdapter),
    [adapters, selectedAdapter],
  );

  return {
    adapters,
    selectedAdapter,
    setSelectedAdapter,
    currentAdapter,
    isRefreshing,
    reload,
  };
}
