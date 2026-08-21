import { useEffect } from 'react';

/**
 * Elements that must keep their native context menu (paste into inputs,
 * copy out of the terminal output, ...).
 */
const NATIVE_MENU_SELECTOR = 'input, textarea, select, [data-native-menu="true"]';

/**
 * Suppresses the context menu on the app chrome only — this is a desktop app,
 * but a blanket `oncontextmenu="return false"` also kills paste-from-menu in
 * the IP fields.
 */
export function useContextMenuGuard(): void {
  useEffect(() => {
    const onContextMenu = (event: MouseEvent) => {
      const target = event.target as Element | null;
      if (target && typeof target.closest === 'function' && target.closest(NATIVE_MENU_SELECTOR)) {
        return;
      }
      event.preventDefault();
    };

    document.addEventListener('contextmenu', onContextMenu);
    return () => document.removeEventListener('contextmenu', onContextMenu);
  }, []);
}
