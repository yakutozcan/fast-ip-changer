import { useCallback, useEffect, useState } from 'react';
import { STORAGE_KEYS, readStored, writeStored } from '../lib/storage';

type ThemePreference = 'light' | 'dark' | null;

const MEDIA_QUERY = '(prefers-color-scheme: dark)';

function readPreference(): ThemePreference {
  const stored = readStored(STORAGE_KEYS.theme);
  return stored === 'light' || stored === 'dark' ? stored : null;
}

function matchSystemDark(): boolean {
  if (typeof window === 'undefined' || !window.matchMedia) return false;
  return window.matchMedia(MEDIA_QUERY).matches;
}

/**
 * Dark mode with an explicit, persisted user choice that falls back to — and
 * keeps following — the OS preference while no choice has been made.
 */
export function useDarkMode() {
  const [preference, setPreference] = useState<ThemePreference>(readPreference);
  const [systemDark, setSystemDark] = useState<boolean>(matchSystemDark);

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return;
    const mql = window.matchMedia(MEDIA_QUERY);
    const onChange = (event: MediaQueryListEvent) => setSystemDark(event.matches);
    mql.addEventListener('change', onChange);
    return () => mql.removeEventListener('change', onChange);
  }, []);

  const darkMode = preference !== null ? preference === 'dark' : systemDark;

  useEffect(() => {
    document.documentElement.classList.toggle('dark', darkMode);
  }, [darkMode]);

  const toggleDarkMode = useCallback(() => {
    const next: Exclude<ThemePreference, null> = darkMode ? 'light' : 'dark';
    writeStored(STORAGE_KEYS.theme, next);
    setPreference(next);
  }, [darkMode]);

  return { darkMode, toggleDarkMode };
}
