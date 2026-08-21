import { useCallback, useEffect, useRef, useState } from 'react';

export type MessageType = 'success' | 'error' | 'info' | '';

export interface AppMessage {
  text: string;
  type: MessageType;
}

export type ShowMessage = (text: string, type: MessageType) => void;

const EMPTY: AppMessage = { text: '', type: '' };

/**
 * Single transient status message with an auto-dismiss timer.
 * The pending timer is always cleared before it is replaced and on unmount.
 */
export function useMessage(timeoutMs = 5000) {
  const [message, setMessage] = useState<AppMessage>(EMPTY);
  const timerRef = useRef<number | null>(null);

  const clearTimer = useCallback(() => {
    if (timerRef.current !== null) {
      window.clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  const showMessage = useCallback<ShowMessage>(
    (text, type) => {
      clearTimer();
      setMessage({ text, type });
      timerRef.current = window.setTimeout(() => {
        timerRef.current = null;
        setMessage(EMPTY);
      }, timeoutMs);
    },
    [clearTimer, timeoutMs],
  );

  /** Message that stays until it is explicitly replaced (e.g. "Uygulanıyor..."). */
  const setPersistentMessage = useCallback<ShowMessage>(
    (text, type) => {
      clearTimer();
      setMessage({ text, type });
    },
    [clearTimer],
  );

  useEffect(() => clearTimer, [clearTimer]);

  return { message, showMessage, setPersistentMessage };
}
